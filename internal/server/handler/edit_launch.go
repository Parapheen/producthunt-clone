package handler

import (
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/Parapheen/ph-clone/internal/app"
	"github.com/Parapheen/ph-clone/internal/domain/launch"
	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/Parapheen/ph-clone/internal/pkg/validation"
	"github.com/google/uuid"
	"github.com/justinas/nosurf"
)

func (h *Handler) GetEditLaunch(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())

	productID := uuid.MustParse(r.PathValue("productID"))
	launchID := uuid.MustParse(r.PathValue("launchID"))

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

	launch, err := h.LaunchService.GetByID(r.Context(), launchID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting launch", slog.Any("error", err))
		h.InternalServerError(w, r, err)
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
		h.InternalServerError(w, r, err)
		return
	}

	err = t.ExecuteTemplate(w, "layout", map[string]interface{}{
		"User":    u,
		"Product": p,
		"Launch":  launch,
		"token":   nosurf.Token(r),
	})
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}
}

func (h *Handler) UpdateLaunch(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())

	// Parse multipart form for potential media uploads
	if err := r.ParseMultipartForm(20 << 20); err != nil { // 20MB
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	productSlug := r.FormValue("product_slug")
	launchID := r.FormValue("launch_id")

	p, err := h.ProductService.GetBySlug(r.Context(), productSlug)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting product", slog.Any("error", err))
		h.InternalServerError(w, r, err)
		return
	}

	if !p.IsOwner(u.ID) {
		http.Error(w, "Вы не автор этого продукта", http.StatusForbidden)
		return
	}

	l, err := h.LaunchService.GetByID(r.Context(), uuid.MustParse(launchID))
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting launch", slog.Any("error", err))
		h.InternalServerError(w, r, err)
		return
	}

	l.Name = r.FormValue("name")
	l.URL = r.FormValue("url")
	l.Tagline = r.FormValue("tagline")
	l.Description = r.FormValue("description")

	launchDate, err := time.Parse("2006-01-02", r.FormValue("launch-date"))

	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error parsing launch date", slog.Any("error", err))
		h.InternalServerError(w, r, err)
		return
	}

	l.LaunchDate = &launchDate

	errors := make([]string, 0)

	// Field-level validation before updating
	v := validation.NewValidator()
	if verr := v.ValidateMultiple(
		v.ValidateString(l.Name, "name", 1, 255, true),
		v.ValidateURL(l.URL, "url", true),
		v.ValidateString(l.Tagline, "tagline", 0, 140, false),
	); verr != nil {
		switch ve := verr.(type) {
		case validation.ValidationErrors:
			for _, e := range ve {
				errors = append(errors, e.Error())
			}
		default:
			errors = append(errors, verr.Error())
		}
	}

	if len(errors) == 0 {
		err = h.LaunchService.Update(r.Context(), l)

		switch err {
		case nil:
			// Optional avatar image (single)
			if fh, ok := r.MultipartForm.File["image"]; ok && len(fh) > 0 {
				f, ferr := fh[0].Open()
				if ferr == nil {
					if fh[0].Size > (10 << 20) {
						errors = append(errors, "Аватар слишком большой (макс 10MB)")
					} else {
						if _, uerr := h.LaunchService.UpdateImage(r.Context(), l.ID, fh[0].Filename, f); uerr != nil {
							h.Logger.ErrorContext(r.Context(), "error saving avatar", slog.Any("error", uerr))
							errors = append(errors, "Ошибка при загрузке аватара")
						}
					}
					f.Close()
				}
			}
			// Handle optional media uploads after successful update
			files := r.MultipartForm.File["media"]
			if len(files) > 0 {
				// Validate media count before processing
				if len(files) > 4 {
					errors = append(errors, "Можно загрузить не более 4 изображений")
				} else {
					// Prepare file uploads for replacement
					uploads := make([]app.FileUpload, 0, len(files))
					for _, fh := range files {
						f, ferr := fh.Open()
						if ferr != nil {
							h.Logger.ErrorContext(r.Context(), "error opening file", slog.Any("error", ferr))
							continue
						}
						uploads = append(uploads, app.FileUpload{
							Filename: fh.Filename,
							Content:  f,
						})
					}

					// Replace all existing media with new uploads
					if err := h.LaunchService.ReplaceMedia(r.Context(), l, uploads); err != nil {
						h.Logger.ErrorContext(r.Context(), "error replacing media", slog.Any("error", err))
						// Check if it's a media limit error
						if err.Error() == "too many media files" {
							errors = append(errors, "Можно загрузить не более 4 изображений")
						} else {
							errors = append(errors, "Ошибка при загрузке изображений")
						}
					}

					// Close all files
					for i := range files {
						if i < len(uploads) {
							if closer, ok := uploads[i].Content.(io.Closer); ok {
								closer.Close()
							}
						}
					}
				}
			}

			if len(errors) == 0 {
				w.Header().Add("HX-Redirect", "/products/"+p.Slug+"/launches/edit")
				return
			}
		case launch.ErrInvalidURL:
			errors = append(errors, "Невалидный URL")
		case launch.ErrLaunchDateInPast:
			errors = append(errors, "Дата запуска не может быть в прошлом")

		default:
			h.Logger.ErrorContext(r.Context(), "error updating launch", slog.Any("error", err))
			errors = append(errors, "Что-то пошло не так. Пожалуйста, попробуйте еще раз.")
		}
	}

	if len(errors) > 0 {
		t, err := template.ParseFiles("views/partials/errors.html")
		if err != nil {
			h.InternalServerError(w, r, err)
			return
		}

		err = t.Execute(w, map[string]interface{}{
			"Errors": errors,
		})
		if err != nil {
			h.InternalServerError(w, r, err)
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
		h.InternalServerError(w, r, err)
		return
	}

	product, err := h.ProductService.GetByID(r.Context(), launch.ProductID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error getting product", slog.Any("error", err))
		h.InternalServerError(w, r, err)
		return
	}

	if !product.IsOwner(user.ID) {
		http.Error(w, "Вы не автор этого продукта", http.StatusForbidden)
		return
	}

	err = h.LaunchService.Delete(r.Context(), launchID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error deleting launch", slog.Any("error", err))
		h.InternalServerError(w, r, err)
		return
	}

	w.Write([]byte(""))
}
