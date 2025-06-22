package handler

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/Parapheen/ph-clone/internal/domain/product"
	"github.com/Parapheen/ph-clone/internal/domain/user"
)

func (h *Handler) GetEditLaunch(w http.ResponseWriter, r *http.Request) {
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

	productSlug := r.PathValue("productSlug")
	launchSlug := r.PathValue("launchSlug")

	p, err := h.ProductService.GetBySlug(r.Context(), productSlug)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting product", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	isOwner := false
	// user is owner of product
	for _, member := range p.Members {
		if member.UserID == user.ID && member.Role == product.Owner {
			isOwner = true
			break
		}
	}

	if !isOwner {
		http.Error(w, "Вы не автор этого продукта", http.StatusForbidden)
		return
	}

	launch, err := h.LaunchService.GetBySlug(r.Context(), launchSlug)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting launch", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles("views/launch.html", "views/header.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, map[string]interface{}{
		"User":    user,
		"Product": p,
		"Launch":  launch,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

