package handler

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/Parapheen/ph-clone/internal/domain/product"
	"github.com/Parapheen/ph-clone/internal/domain/user"
)

func (h *Handler) NewProductForm(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session")
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var user *user.User

	if sessionCookie != nil {
		user, err = h.UserService.GetBySession(r.Context(), sessionCookie.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	t, err := template.ParseFiles("views/new-product.html", "views/header.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, map[string]interface{}{
		"User": user,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) NewProduct(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session")
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var user *user.User

	if sessionCookie != nil {
		user, err = h.UserService.GetBySession(r.Context(), sessionCookie.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if user == nil {
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
		user.ID,
	)

	h.Logger.InfoContext(r.Context(), "created product", slog.Any("name", name), slog.Any("url", url))

	switch err {
	case nil:
		http.Redirect(w, r, "/products/"+p.Slug, http.StatusFound)
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
		return
	}
}
