package middleware

import (
	"net/http"
	"os"

	"github.com/Parapheen/ph-clone/internal/domain/user"
)

func (m *Middleware) AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := user.GetUserFromContext(r.Context())

		if u == nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if u.Email == os.Getenv("ADMIN_EMAIL") {
			next.ServeHTTP(w, r)
			return
		}

		w.WriteHeader(http.StatusForbidden)
	})
}
