package handler

import (
	"log/slog"

	"github.com/Parapheen/ph-clone/internal/app"
	"github.com/jmoiron/sqlx"
)

type Handler struct {
	Logger         *slog.Logger
	DB             *sqlx.DB
	AuthService    AuthService
	UserService    UserService
	ProductService ProductService
	LaunchService  LaunchService
	BlogService    BlogService
	Storage        Storage
	BaseURL        string
	Mailer         app.Mailer
}

func NewHandler(
	logger *slog.Logger,
	db *sqlx.DB,
	authService AuthService,
	userService UserService,
	productService ProductService,
	launchService LaunchService,
	blogService BlogService,
	storage Storage,
	baseURL string,
	mailer app.Mailer,
) *Handler {
	return &Handler{
		Logger:         logger,
		DB:             db,
		AuthService:    authService,
		UserService:    userService,
		ProductService: productService,
		LaunchService:  launchService,
		BlogService:    blogService,
		Storage:        storage,
		BaseURL:        baseURL,
		Mailer:         mailer,
	}
}
