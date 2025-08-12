package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/Parapheen/ph-clone/internal/infra/oauth"
	"golang.org/x/oauth2"
)

type AuthService struct {
	userRepository      user.UserRepository
	yandexOauthProvider *oauth.YandexOauthProvider
    googleOauthProvider *oauth.GoogleOauthProvider
    vkOauthProvider     *oauth.VKOauthProvider
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

func (a *AuthService) GetSocialRedirectURL(provider, state string) string {
    switch provider {
    case "yandex":
        return a.yandexOauthProvider.GetAuthCodeURL(state)
    case "google":
        return a.googleOauthProvider.GetAuthCodeURL(state)
    case "vk":
        return a.vkOauthProvider.GetAuthCodeURL(state)
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
    case "vk":
        token, err = a.vkOauthProvider.Exchange(ctx, code)
        if err != nil {
            return nil, fmt.Errorf("error exchanging code: %w", err)
        }
        account, err = a.vkOauthProvider.GetUserInfo(ctx, token)
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
		err = a.userRepository.Create(ctx, newUser)
		if err != nil {
			return nil, fmt.Errorf("error creating user: %w", err)
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
		return existingUser, nil
	}

	return existingUser, nil
}

func (a *AuthService) Logout(ctx context.Context, session string) error {
	return a.userRepository.DeleteSession(ctx, session)
}
