package handler

import (
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/Parapheen/ph-clone/internal/domain/launch"
	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/google/uuid"
	"github.com/justinas/nosurf"
)

func (h *Handler) GetEditLaunch(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())

	productID := uuid.MustParse(r.PathValue("productID"))
	launchSlug := r.PathValue("launchSlug")

	p, err := h.ProductService.GetByID(r.Context(), productID)

	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting product", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !p.IsOwner(u.ID) {
		http.Error(w, "Вы не автор этого продукта", http.StatusForbidden)
		return
	}

	launch, err := h.LaunchService.GetBySlug(r.Context(), launchSlug)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting launch", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !launch.IsDraft() {
		http.Redirect(w, r, "/products/"+p.Slug, http.StatusFound)
		return
	}

	t, err := template.ParseFiles(
		"views/edit-launch.html",
		"views/partials/launch-state.html",
		"views/layout/layout.html",
		"views/layout/header.html",
		"views/layout/footer.html",
		"views/layout/head.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, map[string]interface{}{
		"User":    u,
		"Product": p,
		"Launch":  launch,
		"token":   nosurf.Token(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UpdateLaunch(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())

	productSlug := r.FormValue("product_slug")
	launchSlug := r.FormValue("launch_slug")

	p, err := h.ProductService.GetBySlug(r.Context(), productSlug)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting product", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !p.IsOwner(u.ID) {
		http.Error(w, "Вы не автор этого продукта", http.StatusForbidden)
		return
	}

	l, err := h.LaunchService.GetBySlug(r.Context(), launchSlug)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting launch", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	l.Name = r.FormValue("name")
	l.URL = r.FormValue("url")
	l.Tagline = r.FormValue("tagline")
	l.Description = r.FormValue("description")

	launchDate, err := time.Parse("2006-01-02", r.FormValue("launch-date"))

	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error parsing launch date", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	l.LaunchDate = &launchDate

	errors := make([]string, 0)

	err = h.LaunchService.Update(r.Context(), l)

	switch err {
	case nil:
		w.Header().Add("HX-Redirect", "/products/"+p.Slug+"/launches")
		return
	case launch.InvalidURLSchemeError, launch.InvalidURL:
		errors = append(errors, "Невалидный URL")
	case launch.LaunchDateInPast:
		errors = append(errors, "Дата запуска не может быть в прошлом")
	default:
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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

func (h *Handler) DeleteLaunch(w http.ResponseWriter, r *http.Request) {
	user := user.GetUserFromContext(r.Context())

	launchID := uuid.MustParse(r.PathValue("launchID"))

	launch, err := h.LaunchService.GetByID(r.Context(), launchID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting launch", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	product, err := h.ProductService.GetByID(r.Context(), launch.ProductID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting product", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !product.IsOwner(user.ID) {
		http.Error(w, "Вы не автор этого продукта", http.StatusForbidden)
		return
	}

	err = h.LaunchService.Delete(r.Context(), launchID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error deleting launch", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(""))
}
