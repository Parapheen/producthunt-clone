package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/justinas/nosurf"
)

func (s *Handler) MyProducts(w http.ResponseWriter, r *http.Request) {
	user := user.GetUserFromContext(r.Context())

	products, err := s.ProductService.GetByOwner(r.Context(), user.ID)
	if err != nil {
		s.Logger.ErrorContext(r.Context(), "error getting launches", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles(
		"views/my-products.html",
		"views/header.html",
		"views/partials/head.html",
		"views/partials/product-edit-card.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, map[string]interface{}{
		"User":     user,
		"Products": products,
		"token":    nosurf.Token(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
