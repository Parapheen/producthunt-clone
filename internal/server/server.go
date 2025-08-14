package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Parapheen/ph-clone/internal/pkg/config"
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
	config  *config.Config
}

func NewServer(h *handler.Handler, m *mw.Middleware, cfg *config.Config) *Server {
	r := chi.NewRouter()

	// Middleware
    r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
    r.Use(m.SessionMiddleware)

    // Custom recoverer that renders our 500 page instead of plain text
    r.Use(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if rec := recover(); rec != nil {
                    // Render generic 500 page; avoid exposing panic detail to user
                    h.InternalServerError(w, r, fmt.Errorf("panic: %v", rec))
                }
            }()
            next.ServeHTTP(w, r)
        })
    })

	// Add security headers
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			if cfg.IsProduction() {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			next.ServeHTTP(w, r)
		})
	})

	// Routes
	r.Get("/", h.Home)
	r.Get("/categories/{categorySlug}", h.CategoryPage)
	r.Get("/promoting", h.PromotingPage)
	r.Get("/rules", h.Rules)
	r.Get("/policy", h.Policy)
	r.Get("/new-product", h.NewProductForm)
	r.Get("/u/{userID}", h.UserProfile)
	r.Get("/u/{userID}/edit", h.EditProfileForm)
    r.Get("/products/{productID}/launches/{launchSlug}/edit", h.GetEditLaunch)
    r.Get("/products/{productSlug}", h.GetProduct)
    // Nested launch page under product slug
    // New index-based route: e.g., /products/foo/launches/1
    r.Get("/products/{productSlug}/launches/{index:[0-9]+}", h.GetProductLaunchByIndex)
    // Only index-based route is supported for public launch pages
    r.Get("/products/{productSlug}/edit", h.EditProductForm)
	r.Get("/products/u/{productID}", h.GetProductByID)
	r.Get("/my/products", h.MyProducts)
	r.Get("/products/{productSlug}/launches", h.ProductLaunches)
	r.Get("/products/{productSlug}/launches/edit", h.EditProductLaunches)
	r.Get("/products/{productSlug}/members", h.ProductMembers)
	r.Get("/products/{productID}/new-launch", h.GetNewLaunch)
	r.Get("/invitations/accept", h.AcceptInvitation)

	// Comments routes (HTMX partials)
	r.Get("/api/launches/{launchID}/comments", h.GetLaunchComments)
	r.Post("/api/launches/{launchID}/comments", h.PostLaunchComment)
	r.Post("/api/launches/{launchID}/comments/{commentID}/reply", h.ReplyLaunchComment)
	r.Post("/api/launches/{launchID}/comments/{commentID}/pin", h.TogglePinComment)

	// Admin routes
	r.Group(func(r chi.Router) {
		r.Use(m.AdminMiddleware)
		r.Get("/admin/moderation/launches", h.ModLaunches)
		r.Post("/api/decline-launch", h.DeclineLaunch)
		r.Post("/api/proceed-launch", h.ProceedLaunch)
	})

    // Auth routes
    r.Get("/auth/yandex", h.YandexAuth)
    r.Get("/auth/yandex/callback", h.YandexAuthCallback)
    r.Get("/auth/google", h.GoogleAuth)
    r.Get("/auth/google/callback", h.GoogleAuthCallback)
    r.Get("/auth/vk", h.VKAuth)
    r.Get("/auth/vk/callback", h.VKAuthCallback)

	// API routes
	r.Get("/api/login", h.LoginModal)
	r.Get("/api/logout", h.Logout)
	r.Post("/api/promote/request", h.RequestPromotion)
	r.Post("/api/new-product", h.NewProduct)
	r.Post("/api/new-launch", h.NewLaunch)
	r.Post("/api/update-launch", h.UpdateLaunch)
	r.Delete("/api/launches/{launchID}", h.DeleteLaunch)
    r.Post("/api/send-launch-to-moderation", h.SendLaunchToModeration)
    r.Post("/api/products/{productID}/profile", h.UpdateProduct)
    r.Post("/api/products/{productID}/invite", h.InviteMember)
	r.Post("/api/users/{userID}/profile", h.UpdateProfile)
	r.Post("/api/launches/{launchID}/upvote", h.ToggleLaunchUpvote)

	// Uploads
	r.Post("/api/users/{userID}/avatar", h.UploadUserAvatar)
	r.Post("/api/products/{productID}/image", h.UploadProductImage)
	r.Post("/api/launches/{launchID}/media", h.UploadLaunchMedia)

	// Static files
	fileServer(r, "/assets")

	// Partials
	r.Get("/api/nav/categories", h.NavCategories)

    // robots.txt and sitemap.xml
    r.Get("/robots.txt", h.Robots)
    r.Get("/sitemap.xml", h.Sitemap)

	// CSRF protection
	csrfHandler := nosurf.New(r)
	csrfHandler.SetBaseCookie(http.Cookie{
		Name:     "ph_csrf",
		Path:     "/",
		MaxAge:   int(cfg.Auth.SessionMaxAge.Seconds()),
		HttpOnly: true,
		Secure:   cfg.IsProduction(),
		SameSite: http.SameSiteLaxMode,
	})
	csrfHandler.SetFailureHandler(http.HandlerFunc(csrfFailureHandler))

	return &Server{
		router:  csrfHandler,
		handler: h,
		config:  cfg,
	}
}

func (s *Server) Run() error {
	addr := ":" + s.config.Server.Port
	
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
		IdleTimeout:  s.config.Server.IdleTimeout,
	}

	// Create channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests in a separate goroutine.
	go func() {
		fmt.Printf("Starting server on %s\n", addr)
		serverErrors <- httpServer.ListenAndServe()
	}()

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return fmt.Errorf("error starting server: %w", err)

	case sig := <-shutdown():
		fmt.Printf("Start shutdown... Signal: %v\n", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Gracefully shutdown the server.
		if err := httpServer.Shutdown(ctx); err != nil {
			fmt.Printf("Could not stop server gracefully: %v\n", err)
			if err := httpServer.Close(); err != nil {
				return fmt.Errorf("could not force close server: %w", err)
			}
		}
	}

	return nil
}

// shutdown returns a channel that will receive shutdown signals.
func shutdown() chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	return ch
}


