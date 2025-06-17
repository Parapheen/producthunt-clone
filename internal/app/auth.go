package app

import (
	"context"
	"fmt"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/Parapheen/ph-clone/internal/infra/oauth"
)

type AuthService struct {
	userRepository      user.UserRepository
	yandexOauthProvider *oauth.YandexOauthProvider
}

func NewAuthService(userRepository user.UserRepository) *AuthService {
	yandexOauthProvider := oauth.NewYandexOauthProvider()

	return &AuthService{
		userRepository:      userRepository,
		yandexOauthProvider: yandexOauthProvider,
	}
}

func (a *AuthService) GetSocialRedirectURL(provider, state string) string {
	if provider != "yandex" {
		return ""
	}

	return a.yandexOauthProvider.GetAuthCodeURL(state)
}

func (a *AuthService) AuthenticateWithSocial(ctx context.Context, provider string, code string) (*user.User, error) {
	if provider != "yandex" {
		return nil, fmt.Errorf("provider %s is not supported", provider)
	}

	token, err := a.yandexOauthProvider.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("error exchanging code: %w", err)
	}

	userInfo, err := a.yandexOauthProvider.GetUserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("error getting user info: %w", err)
	}

	u := user.NewUserFromSocialAccount(userInfo)

	err = a.userRepository.Create(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return u, nil
}

func (a *AuthService) Logout(ctx context.Context, session string) error {
	return a.userRepository.DeleteSession(ctx, session)
}
