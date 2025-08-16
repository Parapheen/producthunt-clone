package handler

import (
	"html/template"
	"net/http"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/Parapheen/ph-clone/internal/pkg/tmpl"
	"github.com/justinas/nosurf"
)

// CategoryPage renders a listing of products for a given category slug.
func (h *Handler) CategoryPage(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())
	slug := r.PathValue("categorySlug")

	cat, err := h.ProductService.GetCategoryBySlug(r.Context(), slug)
	if err != nil {
		// 404 is fine to show as not found (could add custom 404 later)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	products, err := h.ProductService.GetByCategorySlug(r.Context(), slug)
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}

	t, err := template.New("category").Funcs(template.FuncMap{
		"dict": tmpl.Dict,
	}).ParseFiles(
		"views/category.html",
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

	meta := map[string]any{
		"Title":       "Категория — " + cat.Name,
		"Description": "Продукты в категории " + cat.Name,
		"Canonical":   h.BaseURL + "/categories/" + cat.Slug,
	}

	data := map[string]any{
		"User":     u,
		"Category": cat,
		"Products": products,
		"token":    nosurf.Token(r),
		"meta":     meta,
	}
	if err := t.ExecuteTemplate(w, "layout", data); err != nil {
		h.InternalServerError(w, r, err)
		return
	}
}

// NavCategories renders the categories list for the header dropdown
func (h *Handler) NavCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.ProductService.ListCategories(r.Context())
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}
	t, err := template.New("nav-categories").ParseFiles(
		"views/partials/nav-categories.html",
	)
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}
	_ = t.ExecuteTemplate(w, "nav-categories", map[string]any{
		"Categories": cats,
	})
}
