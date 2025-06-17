package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/yandex"
)

type YandexUser struct {
	ID    string `json:"id"`
	Name  string `json:"display_name"`
	Email string `json:"default_email"`
}

type YandexOauthProvider struct {
	config *oauth2.Config
}

func NewYandexOauthProvider() *YandexOauthProvider {
	config := &oauth2.Config{
		ClientID:     os.Getenv("YANDEX_CLIENT_ID"),
		ClientSecret: os.Getenv("YANDEX_CLIENT_SECRET"),
		Endpoint:     yandex.Endpoint,
		RedirectURL:  os.Getenv("YANDEX_REDIRECT_URL"),
		Scopes:       []string{"login:email", "login:info", "login:avatar"},
	}
	return &YandexOauthProvider{
		config: config,
	}
}

func (y *YandexOauthProvider) GetAuthCodeURL(state string) string {
	return y.config.AuthCodeURL(state)
}

func (y *YandexOauthProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := y.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("error exchanging code: %w", err)
	}

	return token, nil
}

func (y *YandexOauthProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*user.SocialAccount, error) {
	resp, err := y.config.Client(ctx, token).Get("https://login.yandex.ru/info")
	if err != nil {
		return nil, fmt.Errorf("error getting user info: %w", err)
	}

	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	var yandexUser YandexUser
	err = json.Unmarshal(body, &yandexUser)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling user info: %w", err)
	}

	return &user.SocialAccount{
		ID:         uuid.New(),
		Provider:   "yandex",
		ProviderID: yandexUser.ID,
		Email:      yandexUser.Email,
		Name:       yandexUser.Name,
	}, nil
}
