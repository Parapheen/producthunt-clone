package handler

import (
	"html/template"
	"io"
	"log/slog"
	"net/http"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/Parapheen/ph-clone/internal/pkg/validation"
	"github.com/google/uuid"
	"github.com/justinas/nosurf"
)

func (h *Handler) EditProfileForm(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())
	if u == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userIDParam := r.PathValue("userID")
	userID, err := uuid.Parse(userIDParam)
	if err != nil || userID != u.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	t, err := template.ParseFiles(
		"views/edit-profile.html",
		"views/layout/layout.html",
		"views/layout/header.html",
		"views/layout/footer.html",
		"views/layout/head.html",
	)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error parsing templates", slog.Any("error", err))
		h.InternalServerError(w, r, err)
		return
	}

	err = t.ExecuteTemplate(w, "layout", map[string]interface{}{
		"User":  u,
		"token": nosurf.Token(r),
	})
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error executing template", slog.Any("error", err))
		h.InternalServerError(w, r, err)
		return
	}
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())
	if u == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userIDParam := r.PathValue("userID")
	userID, err := uuid.Parse(userIDParam)
	if err != nil || userID != u.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Parse multipart to allow optional avatar file along with bio
	const maxAvatarBytes = int64(10 << 20) // 10MB
	if err := r.ParseMultipartForm(maxAvatarBytes); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	bio := r.FormValue("bio")
	// Validate bio length (max 280 chars)
	v := validation.NewValidator()
	if verr := v.ValidateString(bio, "bio", 0, 280, false); verr != nil {
		// Normalize errors to string slice and render
		errors := []string{verr.Error()}
		if ve, ok := verr.(validation.ValidationErrors); ok {
			errors = errors[:0]
			for _, e := range ve {
				errors = append(errors, e.Error())
			}
		}
		h.renderErrors(w, r, errors)
		return
	}

	if err := h.UserService.UpdateBio(r.Context(), u.ID, bio); err != nil {
		h.Logger.ErrorContext(r.Context(), "error updating bio", slog.Any("error", err))
		h.InternalServerError(w, r, err)
		return
	}

	// Optional avatar upload as part of the same form
	if file, header, err := r.FormFile("image"); err == nil && header != nil {
		defer file.Close()
		// Sanity size check
		// If Size is known and exceeds limit, reject
		if header.Size > 0 && header.Size > maxAvatarBytes {
			http.Error(w, "image is too large (max 10MB)", http.StatusBadRequest)
			return
		}
		// Additionally protect downstream by wrapping in LimitedReader
		limited := &io.LimitedReader{R: file, N: maxAvatarBytes + 1}
		if _, err := h.UserService.UpdateAvatar(r.Context(), u.ID, header.Filename, limited); err != nil {
			h.Logger.ErrorContext(r.Context(), "error updating avatar", slog.Any("error", err))
			h.InternalServerError(w, r, err)
			return
		}
		if limited.N <= 0 { // exceeded limit
			http.Error(w, "image is too large (max 10MB)", http.StatusBadRequest)
			return
		}
	}

	// Redirect to profile page
	// If it's an HTMX request, prefer HX-Redirect
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/u/"+u.ID.String())
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/u/"+u.ID.String(), http.StatusSeeOther)
}
