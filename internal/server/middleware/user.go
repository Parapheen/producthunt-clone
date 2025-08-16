package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/Parapheen/ph-clone/internal/domain/user"
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
			next.ServeHTTP(w, r)
		case user.ErrSessionExpired:
			slog.Info("session expired", slog.Any("cookie", cookie))
			http.SetCookie(w, &http.Cookie{
				Name:     sessionCookieName,
				Value:    "",
				HttpOnly: true,
				Path:     "/",
				Expires:  time.Unix(0, 0),
			})
		case user.ErrUserNotFound:
			slog.Info("user not found", slog.Any("cookie", cookie))
			http.SetCookie(w, &http.Cookie{
				Name:     sessionCookieName,
				Value:    "",
				HttpOnly: true,
				Path:     "/",
				Expires:  time.Unix(0, 0),
			})
		default:
			slog.Error("error getting user by session", slog.Any("err", err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	})
}
