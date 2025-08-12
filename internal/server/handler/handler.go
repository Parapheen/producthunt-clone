package handler

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
)

type Handler struct {
    Logger         *slog.Logger
    DB             *sqlx.DB
    AuthService    AuthService
    UserService    UserService
    ProductService ProductService
    LaunchService  LaunchService
    Storage        Storage
}

func NewHandler(
	logger *slog.Logger,
	db *sqlx.DB,
	authService AuthService,
	userService UserService,
	productService ProductService,
	launchService LaunchService,
    storage Storage,
) *Handler {
	return &Handler{
		Logger:         logger,
		DB:             db,
		AuthService:    authService,
		UserService:    userService,
		ProductService: productService,
		LaunchService:  launchService,
        Storage:        storage,
	}
}
