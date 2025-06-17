package user

import (
	"context"

	"golang.org/x/oauth2"
)

type OAuthProvider interface {
	GetAuthCodeURL(state string) string
	Exchange(ctx context.Context, code string) (*oauth2.Token, error)
	GetUserInfo(ctx context.Context, token *oauth2.Token) (*User, error)
}
