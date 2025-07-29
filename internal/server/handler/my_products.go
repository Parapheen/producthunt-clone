package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/Parapheen/ph-clone/internal/domain/launch"
	"github.com/Parapheen/ph-clone/internal/domain/product"
	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/google/uuid"
	"github.com/justinas/nosurf"
)

type ProductView struct {
	Name    string
	Slug    string
	Tagline string
	ID      uuid.UUID

	Launches   []*launch.Launch
	Categories []*product.Category
}

func (s *Handler) MyProducts(w http.ResponseWriter, r *http.Request) {
	user := user.GetUserFromContext(r.Context())

	products, err := s.ProductService.GetByOwner(r.Context(), user.ID)
	if err != nil {
		s.Logger.ErrorContext(r.Context(), "error getting launches", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	productsView := make([]*ProductView, 0, len(products))
	for _, p := range products {
		launches, err := s.LaunchService.GetByProduct(r.Context(), p.ID)
		if err != nil {
			s.Logger.ErrorContext(r.Context(), "error getting launches", slog.Any("error", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		productsView = append(productsView, &ProductView{
			Name:       p.Name,
			Slug:       p.Slug,
			ID:         p.ID,
			Tagline:    p.Tagline,
			Categories: p.Categories,
			Launches:   launches,
		})
	}

	t, err := template.ParseFiles(
		"views/my-products.html",
		"views/layout/layout.html",
		"views/layout/header.html",
		"views/layout/footer.html",
		"views/layout/head.html",
		"views/partials/product-edit-card.html",
		"views/partials/launch-state.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.ExecuteTemplate(w, "layout", map[string]interface{}{
		"User":     user,
		"Products": productsView,
		"token":    nosurf.Token(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
