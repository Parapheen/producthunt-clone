package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/Parapheen/ph-clone/internal/pkg/env"
)

const sessionCookieName = "session"

func (m *Middleware) SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)

		if errors.Is(err, http.ErrNoCookie) {
			next.ServeHTTP(w, r)
			return
		} else if err != nil {
			slog.Error("error getting cookie", slog.Any("err", err))
			next.ServeHTTP(w, r)
			return
		}

		u, err := m.UserService.GetBySession(r.Context(), cookie.Value)

		switch err {
		case nil:
			r = r.WithContext(context.WithValue(r.Context(), user.ContextKeyUser, u))
			http.SetCookie(w, &http.Cookie{
				Name:     sessionCookieName,
				Value:    u.Session.ID.String(),
				Path:     "/",
				HttpOnly: true,
				Secure:   env.IsProduction(),
				Expires:  u.Session.ExpiresAt,
			})
			next.ServeHTTP(w, r)
		case user.ErrSessionExpired:
			http.Redirect(w, r, "/login", http.StatusFound)
		case user.ErrUserNotFound:
			http.Redirect(w, r, "/login", http.StatusFound)
		default:
			slog.Error("error getting user by session", slog.Any("err", err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
}
