package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/justinas/nosurf"
)

// InternalServerError writes a generic 500 error page without exposing internal error details.
func (h *Handler) InternalServerError(w http.ResponseWriter, r *http.Request, err error) {
    if err != nil {
        h.Logger.ErrorContext(r.Context(), "internal server error", slog.Any("error", err))
    }

    // Ensure status code is set once
    w.WriteHeader(http.StatusInternalServerError)

    t, tplErr := template.New("error-500").ParseFiles(
        "views/500.html",
        "views/layout/layout.html",
        "views/layout/header.html",
        "views/layout/footer.html",
        "views/layout/head.html",
    )
    if tplErr != nil {
        // Fallback plain text if templates fail
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        return
    }

    data := map[string]any{
        "token": nosurf.Token(r),
        "meta": map[string]any{
            "Title": "Ошибка 500 — justlaunch",
        },
    }

    _ = t.ExecuteTemplate(w, "layout", data)
}


