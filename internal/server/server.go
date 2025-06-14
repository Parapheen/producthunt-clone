package server

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/yandex"
)

type Server struct {
	router *chi.Mux
}

var yandexConfig = oauth2.Config{
	ClientID:     os.Getenv("YANDEX_CLIENT_ID"),
	ClientSecret: os.Getenv("YANDEX_CLIENT_SECRET"),
	Endpoint:     yandex.Endpoint,
	RedirectURL:  "http://localhost:3333/auth/yandex/callback",
	Scopes:       []string{"email"},
}

func NewServer() *Server {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles("views/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = t.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Get("/api/login", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles("views/partials/auth-modal.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = t.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Get("/auth/yandex", func(w http.ResponseWriter, r *http.Request) {

		state := uuid.NewString()
		http.SetCookie(w, &http.Cookie{
			Name:     "ouath_state",
			Value:    state,
			HttpOnly: true,
		})

		url := yandexConfig.AuthCodeURL(state)

		http.Redirect(w, r, url, http.StatusFound)
	})

	r.Get("/auth/yandex/callback", func(w http.ResponseWriter, r *http.Request) {
		// validate state
		cookie, err := r.Cookie("ouath_state")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		state := r.URL.Query().Get("state")
		if cookie.Value != state {
			http.Error(w, "Invalid state", http.StatusInternalServerError)
			return
		}

		code := r.URL.Query().Get("code")

		token, err := yandexConfig.Exchange(r.Context(), code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		userInfo, err := yandexConfig.Client(r.Context(), token).Get("https://login.yandex.ru/info")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Println(userInfo)

		// clean state cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "ouath_state",
			Value:    "",
			HttpOnly: true,
		})
	})

	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "assets"))
	fileServer(r, "/assets", filesDir)

	return &Server{
		router: r,
	}
}

func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s.router)
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func fileServer(r chi.Router, path string, root http.FileSystem) error {
	if strings.ContainsAny(path, "{}*") {
		return errors.New("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})

	return nil
}
