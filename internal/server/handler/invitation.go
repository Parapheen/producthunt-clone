package handler

import (
	"html/template"
	"net/http"
	"time"

	"github.com/Parapheen/ph-clone/internal/domain/user"
)

// AcceptInvitation handles GET /invitations/accept?token=...
//   - If user is not authenticated: render an informational HTML with guidance to register/login
//     and then click the same link again to complete joining the team.
//   - If user is authenticated: accept invitation and redirect to the product page.
func (h *Handler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	u := user.GetUserFromContext(r.Context())
	if u == nil {
		// Store invite token in a short-lived cookie so we can auto-accept after OAuth login
		http.SetCookie(w, &http.Cookie{
			Name:     inviteTokenCookie,
			Value:    token,
			HttpOnly: true,
			Path:     "/",
			Expires:  time.Now().Add(30 * time.Minute),
		})

        t, err := template.ParseFiles(
			"views/invitations/accept.html",
			"views/layout/layout.html",
			"views/layout/header.html",
			"views/layout/footer.html",
			"views/layout/head.html",
		)
		if err != nil {
            h.InternalServerError(w, r, err)
			return
		}
		if err := t.ExecuteTemplate(w, "layout", map[string]interface{}{}); err != nil {
            h.InternalServerError(w, r, err)
		}
		return
	}

	// Authenticated: accept and redirect to product page
    productID, err := h.ProductService.AcceptInvitation(r.Context(), token, u.ID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
	http.Redirect(w, r, "/products/u/"+productID.String(), http.StatusFound)
}

