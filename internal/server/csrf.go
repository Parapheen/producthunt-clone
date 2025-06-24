package server

import (
	"log/slog"
	"net/http"

	"github.com/justinas/nosurf"
)

func csrfFailureHandler(w http.ResponseWriter, r *http.Request) {
	slog.WarnContext(r.Context(), "CSRF check failed (nosurf)", slog.String("path", r.URL.Path), slog.String("reason", nosurf.Reason(r).Error()))
	http.Error(w, "CSRF validation failed", http.StatusForbidden)
}
