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

	t, err := template.ParseFiles("views/new-product.html", "views/header.html", "views/partials/head.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, map[string]interface{}{
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
		t, err := template.ParseFiles("views/partials/errors.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = t.Execute(w, map[string]interface{}{
			"Errors": errors,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	h.Logger.InfoContext(r.Context(), "creating product", slog.Any("name", name), slog.Any("url", url))

	p, err := h.ProductService.Create(
		r.Context(),
		name,
		url,
		u.ID,
	)

	switch err {
	case nil:
		createdFirstLaunch, errLaunch := h.LaunchService.GetLatestByProduct(r.Context(), p.ID)
		if errLaunch != nil {
			h.Logger.ErrorContext(r.Context(), "error getting latest launch", slog.Any("error", errLaunch))
			errors = append(errors, "Что-то пошло не так. Пожалуйста, попробуйте еще раз.")
		}

		redirectTo := "/products/" + p.Slug + "/launches/" + createdFirstLaunch.Slug + "/edit"
		w.Header().Add("HX-Redirect", redirectTo)
		return
	case product.ProductNameTooLong:
		errors = append(errors, "Название продукта слишком длинное")
	case product.ProductURLTooLong:
		errors = append(errors, "URL продукта слишком длинный")
	case product.InvalidURLSchemeError, product.InvalidURL:
		errors = append(errors, "Невалидный URL")
	case product.ProductNameEmpty:
		errors = append(errors, "Название продукта не может быть пустым")
	case product.ProductURLEmpty:
		errors = append(errors, "URL продукта не может быть пустым")
	default:
		h.Logger.ErrorContext(r.Context(), "error creating product", slog.Any("error", err))
		errors = append(errors, "Что-то пошло не так. Пожалуйста, попробуйте еще раз.")
	}

	if len(errors) > 0 {
		t, err := template.ParseFiles("views/partials/errors.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = t.Execute(w, map[string]interface{}{
			"Errors": errors,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
