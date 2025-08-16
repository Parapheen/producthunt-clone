package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/google/uuid"
)

func (h *Handler) UserProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")

	loggedUser := user.GetUserFromContext(r.Context())

	u, err := h.UserService.GetByID(r.Context(), uuid.MustParse(userID))
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting user", slog.Any("error", err))
		h.InternalServerError(w, r, err)
		return
	}

	products, err := h.ProductService.GetByOwner(r.Context(), u.ID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting products", slog.Any("error", err))
		h.InternalServerError(w, r, err)
		return
	}

	launches, err := h.LaunchService.GetByOwner(r.Context(), u.ID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting launches", slog.Any("error", err))
		h.InternalServerError(w, r, err)
		return
	}

	t, err := template.ParseFiles(
		"views/profile.html",
		"views/layout/layout.html",
		"views/layout/header.html",
		"views/layout/footer.html",
		"views/layout/head.html",
		"views/partials/product-card.html",
	)
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}

	err = t.ExecuteTemplate(w, "layout", map[string]interface{}{
		"User":     loggedUser,
		"Profile":  u,
		"Products": products,
		"Launches": launches,
	})
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}
}
