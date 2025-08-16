package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/Parapheen/ph-clone/internal/domain/launch"
	"github.com/Parapheen/ph-clone/internal/domain/product"
	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/Parapheen/ph-clone/internal/pkg/tmpl"
	"github.com/google/uuid"
	"github.com/justinas/nosurf"
)

type ProductView struct {
	Name     string
	Slug     string
	Tagline  string
	ID       uuid.UUID
	ImageURL string

	Launches   []*launch.Launch
	Categories []*product.Category
}

type MemberProductView struct {
	Product  *product.Product
	Launches []*launch.Launch
}

func (s *Handler) MyProducts(w http.ResponseWriter, r *http.Request) {
	user := user.GetUserFromContext(r.Context())

	// Owned products
	products, err := s.ProductService.GetByOwner(r.Context(), user.ID)
	if err != nil {
		s.Logger.ErrorContext(r.Context(), "error getting launches", slog.Any("error", err))
		s.InternalServerError(w, r, err)
		return
	}

	productsView := make([]*ProductView, 0, len(products))
	for _, p := range products {
		launches, err := s.LaunchService.GetByProduct(r.Context(), p.ID)
		if err != nil {
			s.Logger.ErrorContext(r.Context(), "error getting launches", slog.Any("error", err))
			s.InternalServerError(w, r, err)
			return
		}

		productsView = append(productsView, &ProductView{
			Name:       p.Name,
			Slug:       p.Slug,
			ID:         p.ID,
			Tagline:    p.Tagline,
			ImageURL:   p.ImageURL,
			Categories: p.Categories,
			Launches:   launches,
		})
	}

	// Member (non-owner) products
	memberProducts, err := s.ProductService.GetByMember(r.Context(), user.ID)
	if err != nil {
		s.Logger.ErrorContext(r.Context(), "error getting member products", slog.Any("error", err))
		s.InternalServerError(w, r, err)
		return
	}
	memberViews := make([]*MemberProductView, 0, len(memberProducts))
	for _, p := range memberProducts {
		// Skip those where the user is actually the owner (they're already in the owner list)
		if p.IsOwner(user.ID) {
			continue
		}
		launches, err := s.LaunchService.GetByProduct(r.Context(), p.ID)
		if err != nil {
			s.Logger.ErrorContext(r.Context(), "error getting launches for member product", slog.Any("error", err))
			s.InternalServerError(w, r, err)
			return
		}
		memberViews = append(memberViews, &MemberProductView{
			Product:  p,
			Launches: launches,
		})
	}

	t, err := template.New("my-products").Funcs(template.FuncMap{
		"dict": tmpl.Dict,
	}).ParseFiles(
		"views/my-products.html",
		"views/layout/layout.html",
		"views/layout/header.html",
		"views/layout/footer.html",
		"views/layout/head.html",
		"views/partials/product-edit-card.html",
		"views/partials/launch-state.html",
		"views/partials/product-card.html",
		"views/partials/launch-card.html",
		"views/partials/launch-upvote.html",
	)
	if err != nil {
		s.InternalServerError(w, r, err)
		return
	}

	err = t.ExecuteTemplate(w, "layout", map[string]interface{}{
		"User":           user,
		"OwnerProducts":  productsView,
		"MemberProducts": memberViews,
		"token":          nosurf.Token(r),
	})
	if err != nil {
		s.InternalServerError(w, r, err)
		return
	}
}
