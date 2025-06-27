package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

func (h *Handler) UserProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")

	u, err := h.UserService.GetByID(r.Context(), uuid.MustParse(userID))
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting user", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	products, err := h.ProductService.GetByOwner(r.Context(), u.ID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting products", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	launches, err := h.LaunchService.GetByOwner(r.Context(), u.ID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting launches", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles("views/profile.html", "views/header.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, map[string]interface{}{
		"User":     u,
		"Products": products,
		"Launches": launches,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
