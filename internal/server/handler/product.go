package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/google/uuid"
	"github.com/justinas/nosurf"
)

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())

	productSlug := r.PathValue("productSlug")

	p, err := h.ProductService.GetBySlug(r.Context(), productSlug)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting product", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	launches, err := h.LaunchService.GetByProduct(r.Context(), p.ID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting launches", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t, err := template.New("product.html").Funcs(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}).ParseFiles(
		"views/product.html",
		"views/layout/layout.html",
		"views/layout/header.html",
		"views/layout/footer.html",
		"views/layout/head.html",
		"views/partials/launch-card.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, map[string]interface{}{
		"User":     u,
		"Product":  p,
		"Launches": launches,
		"token":    nosurf.Token(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	productID := r.PathValue("productID")

	p, err := h.ProductService.GetByID(r.Context(), uuid.MustParse(productID))
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting product", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/products/"+p.Slug, http.StatusFound)
}

