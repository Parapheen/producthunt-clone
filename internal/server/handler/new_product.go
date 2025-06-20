package handler

import (
	"errors"
	"html/template"
	"net/http"

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

	// create product
}
