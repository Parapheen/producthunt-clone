package server

import (
	"net/http"

	"github.com/Parapheen/ph-clone/internal/server/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	router *chi.Mux
	hander *handler.Handler
}

func NewServer(h *handler.Handler) *Server {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/", h.Home)

	r.Get("/auth/yandex", h.YandexAuth)
	r.Get("/auth/yandex/callback", h.YandexAuthCallback)

	r.Get("/api/login", h.LoginModal)
	r.Get("/api/logout", h.Logout)

	fileServer(r, "/assets")

	return &Server{
		router: r,
		hander: h,
	}
}

func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s.router)
}
