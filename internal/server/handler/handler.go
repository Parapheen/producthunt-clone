package handler

import "log/slog"

type Handler struct {
	Logger         *slog.Logger
	AuthService    AuthService
	UserService    UserService
	ProductService ProductService
	LaunchService  LaunchService
}

func NewHandler(
	logger *slog.Logger,
	authService AuthService,
	userService UserService,
	productService ProductService,
	launchService LaunchService,
) *Handler {
	return &Handler{
		Logger:         logger,
		AuthService:    authService,
		UserService:    userService,
		ProductService: productService,
		LaunchService:  launchService,
	}
}
