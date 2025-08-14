package handler

import (
	"html/template"
	"io"
	"log/slog"
	"net/http"

	"github.com/Parapheen/ph-clone/internal/domain/product"
	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/Parapheen/ph-clone/internal/pkg/validation"
	"github.com/google/uuid"
	"github.com/justinas/nosurf"
)

func (h *Handler) EditProductForm(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())
	if u == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	productSlug := r.PathValue("productSlug")
	p, err := h.ProductService.GetBySlug(r.Context(), productSlug)
	if err != nil || p == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if !p.IsOwner(u.ID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	t, err := template.ParseFiles(
		"views/edit-product.html",
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
		"User":    u,
		"Product": p,
		"token":   nosurf.Token(r),
	})
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error executing template", slog.Any("error", err))
		h.InternalServerError(w, r, err)
		return
	}
}

func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())
	if u == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	productIDParam := r.PathValue("productID")
	productID, err := uuid.Parse(productIDParam)
	if err != nil {
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}
	p, err := h.ProductService.GetByID(r.Context(), productID)
	if err != nil || p == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if !p.IsOwner(u.ID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	const maxImageBytes = int64(10 << 20) // 10MB
	if err := r.ParseMultipartForm(maxImageBytes); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	tagline := r.FormValue("tagline")
	v := validation.NewValidator()
	if verr := v.ValidateString(tagline, "tagline", 0, 140, false); verr != nil {
		// Render standard error partial
		h.renderErrors(w, r, []string{verr.Error()})
		return
	}
    if err := h.ProductService.UpdateTagline(r.Context(), p.ID, tagline); err != nil {
		h.Logger.ErrorContext(r.Context(), "error updating product tagline", slog.Any("error", err))
        h.InternalServerError(w, r, err)
		return
	}

	if file, header, err := r.FormFile("image"); err == nil && header != nil {
		defer file.Close()
		if header.Size > 0 && header.Size > maxImageBytes {
			http.Error(w, "image is too large (max 10MB)", http.StatusBadRequest)
			return
		}
		limited := &io.LimitedReader{R: file, N: maxImageBytes + 1}
        if _, err := h.ProductService.UpdateImage(r.Context(), p.ID, header.Filename, limited); err != nil {
			h.Logger.ErrorContext(r.Context(), "error updating product image", slog.Any("error", err))
            h.InternalServerError(w, r, err)
			return
		}
		if limited.N <= 0 {
			http.Error(w, "image is too large (max 10MB)", http.StatusBadRequest)
			return
		}
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/products/"+p.Slug)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/products/"+p.Slug, http.StatusSeeOther)
}

// InviteMember handles inviting a member by email to a product
func (h *Handler) InviteMember(w http.ResponseWriter, r *http.Request) {
    u := user.GetUserFromContext(r.Context())
    if u == nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    productIDParam := r.PathValue("productID")
    productID, err := uuid.Parse(productIDParam)
    if err != nil {
        http.Error(w, "invalid product id", http.StatusBadRequest)
        return
    }
    p, err := h.ProductService.GetByID(r.Context(), productID)
    if err != nil || p == nil {
        http.Error(w, "Not found", http.StatusNotFound)
        return
    }
    if !p.IsOwner(u.ID) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    if err := r.ParseForm(); err != nil {
        http.Error(w, "invalid form", http.StatusBadRequest)
        return
    }
    email := r.FormValue("email")
    roleStr := r.FormValue("role")
    role := product.ParseRole(roleStr)
    v := validation.NewValidator()
    if verr := v.ValidateEmail(email, "email", true); verr != nil {
        h.renderErrors(w, r, []string{verr.Error()})
        return
    }
    inv, err := h.ProductService.InviteMember(r.Context(), p.ID, email, role)
    if err != nil {
        h.Logger.ErrorContext(r.Context(), "error inviting member", slog.Any("error", err))
        h.InternalServerError(w, r, err)
        return
    }

    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte(`<div class="p-3 border border-green-200 bg-green-50 text-green-800 rounded-md">` +
        `Письмо-приглашение отправлено на <strong>` + template.HTMLEscapeString(inv.Email) + `</strong>. ` +
        `Попросите пользователя проверить почту и перейти по ссылке для подтверждения.</div>`))
}
