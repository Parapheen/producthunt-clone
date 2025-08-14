package handler

import (
	"html/template"
	"net/http"
	"time"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/google/uuid"
	"github.com/justinas/nosurf"
)

// ToggleLaunchUpvote handles upvote toggle; if unauthorized returns auth modal.
func (h *Handler) ToggleLaunchUpvote(w http.ResponseWriter, r *http.Request) {
    u := user.GetUserFromContext(r.Context())
    if u == nil {
        // Return auth modal HTML for HTMX
        t, err := template.ParseFiles("views/partials/auth-modal.html")
        if err != nil {
            h.InternalServerError(w, r, err)
            return
        }
        w.Header().Set("HX-Retarget", "body")
        w.Header().Set("HX-Reswap", "beforeend")
        if err := t.Execute(w, nil); err != nil {
            h.InternalServerError(w, r, err)
        }
        return
    }

    // Restrict upvoting for accounts younger than 7 days
    // TODO: remove this after testing
    if time.Since(u.CreatedAt) < 1*time.Hour {
        t, err := template.ParseFiles("views/partials/restricted-upvote-modal.html")
        if err != nil {
            h.InternalServerError(w, r, err)
            return
        }
        w.Header().Set("HX-Retarget", "body")
        w.Header().Set("HX-Reswap", "beforeend")
        if err := t.Execute(w, nil); err != nil {
            h.InternalServerError(w, r, err)
        }
        return
    }

    launchIDStr := r.PathValue("launchID")
    launchID, err := uuid.Parse(launchIDStr)
    if err != nil {
        http.Error(w, "invalid launch id", http.StatusBadRequest)
        return
    }

    upvoted, count, err := h.LaunchService.ToggleUpvote(r.Context(), launchID, u.ID)
    if err != nil {
        h.InternalServerError(w, r, err)
        return
    }

    // Re-fetch launch to keep other fields consistent (optional)
    l, err := h.LaunchService.GetByID(r.Context(), launchID)
    if err != nil {
        h.InternalServerError(w, r, err)
        return
    }
    l.Upvotes = count

    t, err := template.New("launch-upvote").ParseFiles("views/partials/launch-upvote.html")
    if err != nil {
        h.InternalServerError(w, r, err)
        return
    }

    withText := r.FormValue("withText") == "true"

    err = t.ExecuteTemplate(w, "launch-upvote", map[string]interface{}{
        "Launch": l,
        "token":  nosurf.Token(r),
        "Upvoted": upvoted,
        "WithText": withText,
    })
    if err != nil {
        h.InternalServerError(w, r, err)
        return
    }
}