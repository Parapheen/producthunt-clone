package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/Parapheen/ph-clone/internal/domain/launch"
	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/google/uuid"
	"github.com/justinas/nosurf"
)

func (h *Handler) GetNewLaunch(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())

	productID := uuid.MustParse(r.PathValue("productID"))

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

	t, err := template.ParseFiles(
		"views/new-launch.html",
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
		"token":   nosurf.Token(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) NewLaunch(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())

	if u == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	errors := make([]string, 0)

	name := r.FormValue("name")
	url := r.FormValue("url")
	productID := uuid.MustParse(r.FormValue("product_id"))

	launch := launch.NewLaunch(productID, name, url)

	launch.Tagline = r.FormValue("tagline")
	launch.Description = r.FormValue("description")

	err := h.LaunchService.Create(
		r.Context(),
		launch,
	)

	switch err {
	case nil:
		redirectTo := "/products/" + productID.String() + "/launches/"
		w.Header().Add("HX-Redirect", redirectTo)
		return
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

