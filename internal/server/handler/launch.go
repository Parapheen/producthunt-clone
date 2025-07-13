package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/goforj/godump"
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
		"views/header.html",
		"views/partials/head.html",
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

	godump.Dump(p)

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

	launch.Name = r.FormValue("name")
	launch.URL = r.FormValue("url")
	launch.Tagline = r.FormValue("tagline")
	launch.Description = r.FormValue("description")

	err = h.LaunchService.Update(r.Context(), launch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Redirect", "/products/"+p.Slug)
}

func (h *Handler) DeleteLaunch(w http.ResponseWriter, r *http.Request) {
	launchID := uuid.MustParse(r.PathValue("launchID"))

	// TODO: check if user is owner of the launch's product

	err := h.LaunchService.Delete(r.Context(), launchID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error deleting launch", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Redirect", "/my/launches")
}
