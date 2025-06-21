package handler

import "log/slog"

type Handler struct {
	Logger         *slog.Logger
	AuthService    AuthService
	UserService    UserService
	ProductService ProductService
}

func NewHandler(
	logger *slog.Logger,
	authService AuthService,
	userService UserService,
	ProductService ProductService,
) *Handler {
	return &Handler{
		Logger:         logger,
		AuthService:    authService,
		UserService:    userService,
		ProductService: ProductService,
	}
}
