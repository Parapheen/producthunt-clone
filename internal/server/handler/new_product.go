package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/Parapheen/ph-clone/internal/domain/product"
	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/justinas/nosurf"
)

func (h *Handler) NewProductForm(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())

	t, err := template.ParseFiles(
		"views/new-product.html",
		"views/partials/select-categories.html",
		"views/layout/layout.html",
		"views/layout/header.html",
		"views/layout/footer.html",
		"views/layout/head.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.ExecuteTemplate(w, "layout", map[string]interface{}{
		"User":  u,
		"token": nosurf.Token(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) NewProduct(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())

	if u == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	errors := make([]string, 0)

	name := r.FormValue("name")
	url := r.FormValue("url")
	tagline := r.FormValue("tagline")

	nameExists, err := h.ProductService.NameExists(r.Context(), name)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error checking if product name exists", slog.Any("error", err))
		errors = append(errors, "Что-то пошло не так. Пожалуйста, попробуйте еще раз.")
	}

	urlExists, err := h.ProductService.URLExists(r.Context(), url)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error checking if product url exists", slog.Any("error", err))
		errors = append(errors, "Что-то пошло не так. Пожалуйста, попробуйте еще раз.")
	}

	if nameExists {
		errors = append(errors, "Продукт с таким названием уже существует")
	}

	if urlExists {
		errors = append(errors, "Продукт с таким URL уже существует")
	}

	if len(errors) > 0 {
		h.renderErrors(w, errors)
		return
	}

	h.Logger.InfoContext(r.Context(), "creating product", slog.Any("name", name), slog.Any("url", url))

	categories := make([]*product.Category, 0)
	for _, category := range r.PostForm["categories"] {
		c, err := h.ProductService.GetCategoryBySlug(r.Context(), category)
		if err != nil {
			h.Logger.ErrorContext(r.Context(), "error getting category", slog.Any("error", err))
			errors = append(errors, "Что-то пошло не так. Пожалуйста, попробуйте еще раз.")
			continue
		}
		categories = append(categories, c)
	}

	p := product.NewProduct(name, url, tagline, categories, u.ID)

	err = h.ProductService.Create(
		r.Context(),
		p,
	)

	switch err {
	case nil:
		createdFirstLaunch, errLaunch := h.LaunchService.GetLatestByProduct(r.Context(), p.ID)
		if errLaunch != nil {
			h.Logger.ErrorContext(r.Context(), "error getting latest launch", slog.Any("error", errLaunch))
			errors = append(errors, "Что-то пошло не так. Пожалуйста, попробуйте еще раз.")
		}

		redirectTo := "/products/" + p.ID.String() + "/launches/" + createdFirstLaunch.Slug + "/edit"
		w.Header().Add("HX-Redirect", redirectTo)
		return
	case product.ErrProductNameTooLong:
		errors = append(errors, "Название продукта слишком длинное")
	case product.ErrProductURLTooLong:
		errors = append(errors, "URL продукта слишком длинный")
	case product.ErrInvalidURLScheme, product.ErrInvalidURL:
		errors = append(errors, "Невалидный URL")
	case product.ErrProductNameEmpty:
		errors = append(errors, "Название продукта не может быть пустым")
	case product.ErrProductURLEmpty:
		errors = append(errors, "URL продукта не может быть пустым")
	case product.ErrCategoryNotFound:
		errors = append(errors, "Категория не найдена")
	case product.ErrNoCategories:
		errors = append(errors, "Необходимо добавить хотя бы одну категорию")
	case product.ErrTooManyCategories:
		errors = append(errors, "Не более 3 категорий")
	default:
		h.Logger.ErrorContext(r.Context(), "error creating product", slog.Any("error", err))
		errors = append(errors, "Что-то пошло не так. Пожалуйста, попробуйте еще раз.")
	}

	if len(errors) > 0 {
		h.renderErrors(w, errors)
		return
	}
}

func (h *Handler) renderErrors(w http.ResponseWriter, errors []string) {
	t, err := template.ParseFiles("views/partials/errors.html")
	if err != nil {
		h.Logger.Error("failed to parse errors template", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, map[string]interface{}{
		"Errors": errors,
	})
	if err != nil {
		h.Logger.Error("failed to execute errors template", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
