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
	"golang.org/x/oauth2/google"
)

type GoogleUser struct {
	ID      string `json:"sub"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
}

type GoogleOauthProvider struct {
	config *oauth2.Config
}

func NewGoogleOauthProvider() *GoogleOauthProvider {
	config := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Endpoint:     google.Endpoint,
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes:       []string{"openid", "email", "profile"},
	}
	return &GoogleOauthProvider{config: config}
}

func (g *GoogleOauthProvider) GetAuthCodeURL(state string) string {
	// Access type offline to potentially get refresh token in future if needed
	return g.config.AuthCodeURL(state)
}

func (g *GoogleOauthProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("error exchanging code: %w", err)
	}
	return token, nil
}

func (g *GoogleOauthProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*user.SocialAccount, error) {
	// Use Google OAuth2 userinfo endpoint
	resp, err := g.config.Client(ctx, token).Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, fmt.Errorf("error getting user info: %w", err)
	}

	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	var gu GoogleUser
	if err := json.Unmarshal(body, &gu); err != nil {
		return nil, fmt.Errorf("error unmarshaling user info: %w", err)
	}

	return &user.SocialAccount{
		ID:         uuid.New(),
		Provider:   "google",
		ProviderID: gu.ID,
		Email:      gu.Email,
		Name:       gu.Name,
		AvatarURL:  gu.Picture,
	}, nil
}
