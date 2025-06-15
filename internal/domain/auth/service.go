package auth

import (
	"os"

	"github.com/jmoiron/sqlx"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/yandex"
)

type Service struct {
	db *sqlx.DB

	SessionService SessionService
	UserService    UserService
	AccountService AccountService

	YandexConfig oauth2.Config
}

func NewService(userService UserService, accountService AccountService, sessionService SessionService) *Service {
	yandexConfig := oauth2.Config{
		ClientID:     os.Getenv("YANDEX_CLIENT_ID"),
		ClientSecret: os.Getenv("YANDEX_CLIENT_SECRET"),
		Endpoint:     yandex.Endpoint,
		RedirectURL:  os.Getenv("YANDEX_REDIRECT_URL"),
		Scopes:       []string{"login:email", "login:info", "login:avatar"},
	}

	return &Service{
		YandexConfig:   yandexConfig,
		UserService:    userService,
		AccountService: accountService,
		SessionService: sessionService,
	}
}
