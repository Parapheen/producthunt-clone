package server

import (
	"net/http"

	"github.com/Parapheen/ph-clone/internal/pkg/env"
	"github.com/Parapheen/ph-clone/internal/server/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/justinas/nosurf"

	mw "github.com/Parapheen/ph-clone/internal/server/middleware"
)

type Server struct {
	router  http.Handler
	handler *handler.Handler
	csrf    *nosurf.CSRFHandler
}

func NewServer(h *handler.Handler, m *mw.Middleware) *Server {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/", h.Home)
	r.Get("/new-product", h.NewProductForm)
	r.Get("/products/{productSlug}/launches/{launchSlug}/edit", h.GetEditLaunch)

	r.Get("/auth/yandex", h.YandexAuth)
	r.Get("/auth/yandex/callback", h.YandexAuthCallback)

	r.Get("/api/login", h.LoginModal)
	r.Get("/api/logout", h.Logout)
	r.Post("/api/new-product", h.NewProduct)
	r.Post("/api/update-launch", h.UpdateLaunch)

	fileServer(r, "/assets")

	csrfHandler := nosurf.New(r)

	csrfHandler.SetBaseCookie(http.Cookie{
		Name:     "ph_csrf",
		Path:     "/",
		MaxAge:   12 * 3600, // Example: 12 hours
		HttpOnly: true,
		Secure:   env.IsProduction(),
		SameSite: http.SameSiteLaxMode,
	})

	// Set the custom failure handler
	csrfHandler.SetFailureHandler(http.HandlerFunc(csrfFailureHandler))

	return &Server{
		router:  csrfHandler,
		handler: h,
	}
}

func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s.router)
}
