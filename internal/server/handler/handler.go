package handler

type Handler struct {
	AuthService AuthService
}

func NewHandler(authService AuthService) *Handler {
	return &Handler{
		AuthService: authService,
	}
}
