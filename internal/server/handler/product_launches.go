package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/Parapheen/ph-clone/internal/pkg/tmpl"
	"github.com/justinas/nosurf"
)

func (s *Handler) ProductLaunches(w http.ResponseWriter, r *http.Request) {
	user := user.GetUserFromContext(r.Context())

	productSlug := r.PathValue("productSlug")

	p, err := s.ProductService.GetBySlug(r.Context(), productSlug)
	if err != nil {
		s.Logger.ErrorContext(r.Context(), "error getting product", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !p.IsOwner(user.ID) {
		http.Error(w, "Вы не автор этого продукта", http.StatusForbidden)
		return
	}

	launches, err := s.LaunchService.GetByProduct(r.Context(), p.ID)
	if err != nil {
		s.Logger.ErrorContext(r.Context(), "error getting launches", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t, err := template.New("product-launches.html").
		Funcs(template.FuncMap{
			"dict": tmpl.Dict,
		}).
		ParseFiles(
			"views/product-launches.html",
			"views/partials/launch-state.html",
			"views/layout/layout.html",
			"views/layout/header.html",
			"views/layout/footer.html",
			"views/layout/head.html",
			"views/partials/launch-edit-card.html",
		)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, map[string]interface{}{
		"User":     user,
		"Launches": launches,
		"Product":  p,
		"token":    nosurf.Token(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
