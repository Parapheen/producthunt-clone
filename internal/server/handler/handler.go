package handler

type Handler struct {
	AuthService AuthService
	UserService UserService
}

func NewHandler(authService AuthService, userService UserService) *Handler {
	return &Handler{
		AuthService: authService,
		UserService: userService,
	}
}
