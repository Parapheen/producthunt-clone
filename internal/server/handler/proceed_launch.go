package handler

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Parapheen/ph-clone/internal/domain/product"
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
        h.InternalServerError(w, r, err)
        return
    }

	if !p.IsOwner(u.ID) {
		http.Error(w, "Вы не автор этого продукта", http.StatusForbidden)
		return
	}

    l, err := h.LaunchService.GetByID(r.Context(), launchID)
    if err != nil {
        h.Logger.ErrorContext(r.Context(), "error getting launch", slog.Any("error", err))
        h.InternalServerError(w, r, err)
        return
    }

	l.ProceedState()

    err = h.LaunchService.Update(r.Context(), l)
    if err != nil {
        h.Logger.ErrorContext(r.Context(), "error updating launch", slog.Any("error", err))
        h.InternalServerError(w, r, err)
        return
    }

	t, err := template.ParseFiles(
		"views/partials/launch-edit-card.html",
		"views/partials/launch-state.html",
	)
    if err != nil {
        h.InternalServerError(w, r, err)
        return
    }

	err = t.ExecuteTemplate(w, "launch-edit-card", map[string]interface{}{
		"Launch": l,
		"token":  nosurf.Token(r),
	})

    if err != nil {
        h.InternalServerError(w, r, err)
        return
    }
}

func (h *Handler) ProceedLaunch(w http.ResponseWriter, r *http.Request) {
	launchID := uuid.MustParse(r.FormValue("launch_id"))

	l, err := h.LaunchService.GetByID(r.Context(), launchID)
    if err != nil {
        h.Logger.ErrorContext(r.Context(), "error getting launch", slog.Any("error", err))
        h.InternalServerError(w, r, err)
        return
    }

	l.ProceedState()

	err = h.LaunchService.Update(r.Context(), l)
    if err != nil {
        h.Logger.ErrorContext(r.Context(), "error updating launch", slog.Any("error", err))
        h.InternalServerError(w, r, err)
        return
    }

    // Send email notification if published
    if h.Mailer != nil && l.IsPublished() {
        // Load product to get members (owners) to email
        p, pErr := h.ProductService.GetByID(r.Context(), l.ProductID)
        if pErr == nil {
            ownerIDs := make([]uuid.UUID, 0)
            for _, m := range p.Members {
                if m.Role == product.Owner {
                    ownerIDs = append(ownerIDs, m.UserID)
                }
            }
            if len(ownerIDs) > 0 {
                users, uErr := h.UserService.GetByIDs(r.Context(), ownerIDs)
                if uErr == nil {
                    // Build link to the published launch page using index route
                    index, _ := h.LaunchService.GetIndexByProductAndLaunchID(r.Context(), p.ID, l.ID)
                    productURL := h.BaseURL + "/products/" + p.Slug + "/launches/" + strconv.Itoa(index)
                    data := map[string]any{
                        "ProductName": p.Name,
                        "LaunchName":  l.Name,
                        "LaunchURL":   productURL,
                    }
                    for _, usr := range users {
                        _ = h.Mailer.Send(r.Context(), usr.Email, "launch_accepted.html", data)
                    }
                }
            }
        }
    }

    t, err := template.ParseFiles(
        "views/partials/toast.html",
    )
    if err != nil {
        h.InternalServerError(w, r, err)
        return
    }

    err = t.ExecuteTemplate(w, "toast", map[string]interface{}{
        "Title":   "Запуск опубликован",
        "Message": "Запуск опубликован и доступен для просмотра",
    })

    if err != nil {
        h.InternalServerError(w, r, err)
        return
    }
}

func (h *Handler) DeclineLaunch(w http.ResponseWriter, r *http.Request) {
	launchID := uuid.MustParse(r.FormValue("launch_id"))

    l, err := h.LaunchService.GetByID(r.Context(), launchID)
    if err != nil {
        h.Logger.ErrorContext(r.Context(), "error getting launch", slog.Any("error", err))
        h.InternalServerError(w, r, err)
        return
    }

	l.Decline()

    err = h.LaunchService.Update(r.Context(), l)
    if err != nil {
        h.Logger.ErrorContext(r.Context(), "error updating launch", slog.Any("error", err))
        h.InternalServerError(w, r, err)
        return
    }

    // Send email notification to owners about decline
    if h.Mailer != nil && l.IsDeclined() {
        p, pErr := h.ProductService.GetByID(r.Context(), l.ProductID)
        if pErr == nil {
            ownerIDs := make([]uuid.UUID, 0)
            for _, m := range p.Members {
                if m.Role == product.Owner {
                    ownerIDs = append(ownerIDs, m.UserID)
                }
            }
            if len(ownerIDs) > 0 {
                users, uErr := h.UserService.GetByIDs(r.Context(), ownerIDs)
                if uErr == nil {
                    // Link to edit page to fix and resubmit
                    productURL := h.BaseURL + "/products/" + p.ID.String() + "/launches/" + l.Slug + "/edit"
                    data := map[string]any{
                        "ProductName": p.Name,
                        "LaunchName":  l.Name,
                        "LaunchURL":   productURL,
                    }
                    for _, usr := range users {
                        _ = h.Mailer.Send(r.Context(), usr.Email, "launch_declined.html", data)
                    }
                }
            }
        }
    }

    t, err := template.ParseFiles(
        "views/partials/toast.html",
    )
    if err != nil {
        h.InternalServerError(w, r, err)
        return
    }

    err = t.ExecuteTemplate(w, "toast", map[string]interface{}{
        "Title":   "Запуск отклонён",
        "Message": "Запуск отклонён модератором",
    })

    if err != nil {
        h.InternalServerError(w, r, err)
        return
    }
}
