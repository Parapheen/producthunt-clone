package app

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/Parapheen/ph-clone/internal/infra/oauth"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type AuthService struct {
	userRepository      user.UserRepository
	yandexOauthProvider *oauth.YandexOauthProvider
	googleOauthProvider *oauth.GoogleOauthProvider
	vkOauthProvider     *oauth.VKOauthProvider
	storage             Storage
	mailer              Mailer
	baseURL             string
}

func NewAuthService(userRepository user.UserRepository) *AuthService {
	yandexOauthProvider := oauth.NewYandexOauthProvider()
	googleOauthProvider := oauth.NewGoogleOauthProvider()
	vkOauthProvider := oauth.NewVKOauthProvider()

	return &AuthService{
		userRepository:      userRepository,
		yandexOauthProvider: yandexOauthProvider,
		googleOauthProvider: googleOauthProvider,
		vkOauthProvider:     vkOauthProvider,
	}
}

// WithStorage wires storage so we can persist provider avatars on first login
func (a *AuthService) WithStorage(storage Storage) *AuthService {
	a.storage = storage
	return a
}

// WithMailer wires a Mailer to send transactional auth-related emails
func (a *AuthService) WithMailer(mailer Mailer) *AuthService {
	a.mailer = mailer
	return a
}

// WithBaseURL sets the public base URL used to build absolute links in emails
func (a *AuthService) WithBaseURL(baseURL string) *AuthService {
	a.baseURL = baseURL
	return a
}

func (a *AuthService) GetSocialRedirectURL(provider, state string) string {
	switch provider {
	case "yandex":
		return a.yandexOauthProvider.GetAuthCodeURL(state)
	case "google":
		return a.googleOauthProvider.GetAuthCodeURL(state)
	default:
		return ""
	}
}

func (a *AuthService) AuthenticateWithSocial(ctx context.Context, provider string, code string) (*user.User, error) {
	var (
		token   *oauth2.Token
		account *user.SocialAccount
		err     error
	)

	switch provider {
	case "yandex":
		token, err = a.yandexOauthProvider.Exchange(ctx, code)
		if err != nil {
			return nil, fmt.Errorf("error exchanging code: %w", err)
		}
		account, err = a.yandexOauthProvider.GetUserInfo(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("error getting user info: %w", err)
		}
	case "google":
		token, err = a.googleOauthProvider.Exchange(ctx, code)
		if err != nil {
			return nil, fmt.Errorf("error exchanging code: %w", err)
		}
		account, err = a.googleOauthProvider.GetUserInfo(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("error getting user info: %w", err)
		}
	default:
		return nil, fmt.Errorf("provider %s is not supported", provider)
	}

	existingUser, err := a.userRepository.GetByProvider(ctx, account.Provider, account.ProviderID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	isNewUser := existingUser == nil

	if isNewUser {
		newUser := user.NewUserFromSocialAccount(account)
		// Try to persist provider avatar if provided
		if a.storage != nil && account.AvatarURL != "" {
			if url, err := a.fetchAndSaveAvatar(ctx, newUser.ID, account.AvatarURL); err == nil {
				newUser.AvatarURL = url
				_ = a.userRepository.UpdateAvatarURL(ctx, newUser.ID, url)
			}
		}
		err = a.userRepository.Create(ctx, newUser)
		if err != nil {
			return nil, fmt.Errorf("error creating user: %w", err)
		}
		// Fire-and-forget welcome email
		if a.mailer != nil {
			go func(u *user.User) {
				// Build links
				newProductPath := "/new-product"
				principlesPath := "/values"
				promotingPath := "/promoting"
				newProductURL := newProductPath
				principlesURL := principlesPath
				promotingURL := promotingPath
				if a.baseURL != "" {
					newProductURL = a.baseURL + newProductPath
					principlesURL = a.baseURL + principlesPath
					promotingURL = a.baseURL + promotingPath
				}
				data := map[string]any{
					"UserName":      u.Name,
					"NewProductURL": newProductURL,
					"PrinciplesURL": principlesURL,
					"PromotingURL":  promotingURL,
				}
				_ = a.mailer.Send(context.WithoutCancel(ctx), u.Email, "welcome.html", data)
			}(newUser)
		}
		return newUser, nil
	}

	if existingUser.Session == nil {
		existingUser.Session = user.NewSession()
		err = a.userRepository.CreateSession(ctx, existingUser)
		if err != nil {
			return nil, fmt.Errorf("error refreshing session: %w", err)
		}
		return existingUser, nil
	}

	if existingUser.Session.IsExpired() {
		existingUser.Session.Refresh()
		err = a.userRepository.RefreshSession(ctx, existingUser.Session)
		if err != nil {
			return nil, fmt.Errorf("error refreshing session: %w", err)
		}
		// fall through to possibly enrich avatar before returning
	}

	// If user has no avatar saved yet but provider supplies one, fetch and persist once
	if existingUser.AvatarURL == "" && account.AvatarURL != "" && a.storage != nil {
		if url, err := a.fetchAndSaveAvatar(ctx, existingUser.ID, account.AvatarURL); err == nil {
			existingUser.AvatarURL = url
			_ = a.userRepository.UpdateAvatarURL(ctx, existingUser.ID, url)
		}
	}

	return existingUser, nil
}

func (a *AuthService) Logout(ctx context.Context, session string) error {
	return a.userRepository.DeleteSession(ctx, session)
}

// fetchAndSaveAvatar downloads the given URL and stores it via configured storage
func (a *AuthService) fetchAndSaveAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) (string, error) {
	if a.storage == nil || avatarURL == "" {
		return "", fmt.Errorf("storage or url missing")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, avatarURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("failed to fetch avatar: status %d", resp.StatusCode)
	}
	var reader io.Reader = resp.Body
	// Use a generic filename; storage will hash/rename appropriately
	return a.storage.Save(ctx, fmt.Sprintf("users/%s/avatars", userID.String()), "provider_avatar", reader)
}

// GetVKAuthURL builds VK ID authorization URL with provided PKCE code challenge
func (a *AuthService) GetVKAuthURL(state string, codeChallenge string) string {
	return a.vkOauthProvider.BuildAuthURL(state, codeChallenge)
}

// AuthenticateWithVK exchanges code using PKCE and device_id, fetches user info, and creates/loads user
func (a *AuthService) AuthenticateWithVK(ctx context.Context, code, codeVerifier, deviceID, state string) (*user.User, error) {
	tok, err := a.vkOauthProvider.ExchangeCode(ctx, code, codeVerifier, deviceID, state)
	if err != nil {
		return nil, fmt.Errorf("error exchanging code: %w", err)
	}
	account, err := a.vkOauthProvider.GetUserInfo(ctx, tok.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("error getting user info: %w", err)
	}

	existingUser, err := a.userRepository.GetByProvider(ctx, account.Provider, account.ProviderID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	isNewUser := existingUser == nil
	if isNewUser {
		newUser := user.NewUserFromSocialAccount(account)
		if a.storage != nil && account.AvatarURL != "" {
			if url, err := a.fetchAndSaveAvatar(ctx, newUser.ID, account.AvatarURL); err == nil {
				newUser.AvatarURL = url
				_ = a.userRepository.UpdateAvatarURL(ctx, newUser.ID, url)
			}
		}
		err = a.userRepository.Create(ctx, newUser)
		if err != nil {
			return nil, fmt.Errorf("error creating user: %w", err)
		}
		if a.mailer != nil {
			go func(u *user.User) {
				newProductPath := "/new-product"
				principlesPath := "/values"
				promotingPath := "/promoting"
				newProductURL := newProductPath
				principlesURL := principlesPath
				promotingURL := promotingPath
				if a.baseURL != "" {
					newProductURL = a.baseURL + newProductPath
					principlesURL = a.baseURL + principlesPath
					promotingURL = a.baseURL + promotingPath
				}
				data := map[string]any{
					"UserName":      u.Name,
					"NewProductURL": newProductURL,
					"PrinciplesURL": principlesURL,
					"PromotingURL":  promotingURL,
				}
				_ = a.mailer.Send(context.WithoutCancel(ctx), u.Email, "welcome.html", data)
			}(newUser)
		}
		return newUser, nil
	}

	if existingUser.Session == nil {
		existingUser.Session = user.NewSession()
		err = a.userRepository.CreateSession(ctx, existingUser)
		if err != nil {
			return nil, fmt.Errorf("error refreshing session: %w", err)
		}
		return existingUser, nil
	}

	if existingUser.Session.IsExpired() {
		existingUser.Session.Refresh()
		err = a.userRepository.RefreshSession(ctx, existingUser.Session)
		if err != nil {
			return nil, fmt.Errorf("error refreshing session: %w", err)
		}
	}

	if existingUser.AvatarURL == "" && account.AvatarURL != "" && a.storage != nil {
		if url, err := a.fetchAndSaveAvatar(ctx, existingUser.ID, account.AvatarURL); err == nil {
			existingUser.AvatarURL = url
			_ = a.userRepository.UpdateAvatarURL(ctx, existingUser.ID, url)
		}
	}

	return existingUser, nil
}
