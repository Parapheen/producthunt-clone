package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/google/uuid"
	"github.com/justinas/nosurf"
)

func (h *Handler) SendLaunchToModeration(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())

	productID := uuid.MustParse(r.FormValue("product_id"))
	launchID := uuid.MustParse(r.FormValue("launch_id"))

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

	l, err := h.LaunchService.GetByID(r.Context(), launchID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting launch", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	l.ProceedState()

	err = h.LaunchService.Update(r.Context(), l)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error updating launch", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles(
		"views/partials/launch-edit-card.html",
		"views/partials/launch-state.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.ExecuteTemplate(w, "launch-edit-card", map[string]interface{}{
		"Launch": l,
		"token":  nosurf.Token(r),
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) ProceedLaunch(w http.ResponseWriter, r *http.Request) {
	launchID := uuid.MustParse(r.FormValue("launch_id"))

	l, err := h.LaunchService.GetByID(r.Context(), launchID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting launch", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	l.ProceedState()

	err = h.LaunchService.Update(r.Context(), l)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error updating launch", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles(
		"views/partials/toast.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.ExecuteTemplate(w, "toast", map[string]interface{}{
		"Title":   "Запуск опубликован",
		"Message": "Запуск опубликован и доступен для просмотра",
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) DeclineLaunch(w http.ResponseWriter, r *http.Request) {
	launchID := uuid.MustParse(r.FormValue("launch_id"))

	l, err := h.LaunchService.GetByID(r.Context(), launchID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting launch", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	l.Decline()

	err = h.LaunchService.Update(r.Context(), l)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error updating launch", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles(
		"views/partials/toast.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.ExecuteTemplate(w, "toast", map[string]interface{}{
		"Title":   "Запуск опубликован",
		"Message": "Запуск опубликован и доступен для просмотра",
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
