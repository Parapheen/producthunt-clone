package handler

import (
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/Parapheen/ph-clone/internal/domain/launch"
	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/Parapheen/ph-clone/internal/pkg/validation"
	"github.com/google/uuid"
	"github.com/justinas/nosurf"
)

func (h *Handler) GetNewLaunch(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())

	productID := uuid.MustParse(r.PathValue("productID"))

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

	t, err := template.ParseFiles(
		"views/new-launch.html",
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
		"token":   nosurf.Token(r),
	})
	if err != nil {
		h.InternalServerError(w, r, err)
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

	// Multipart for optional media
	if err := r.ParseMultipartForm(20 << 20); err != nil { // 20MB
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	url := r.FormValue("url")
	tagline := r.FormValue("tagline")
	description := r.FormValue("description")
	launchDate, err := time.Parse("2006-01-02", r.FormValue("launch-date"))
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "error parsing launch date", slog.Any("error", err))
		h.InternalServerError(w, r, err)
		return
	}

	productID := uuid.MustParse(r.FormValue("product_id"))

	// Field-level validation before creating domain entity
	v := validation.NewValidator()
	if verr := v.ValidateMultiple(
		v.ValidateString(name, "name", 1, 255, true),
		v.ValidateURL(url, "url", true),
		v.ValidateString(tagline, "tagline", 0, 140, false),
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

	l := launch.NewLaunch(productID, name, url)

	l.Tagline = tagline
	l.Description = description
	l.LaunchDate = &launchDate

	err = h.LaunchService.Create(
		r.Context(),
		l,
	)

	switch err {
	case nil:
		// Optional avatar image (single)
		if fh, ok := r.MultipartForm.File["image"]; ok && len(fh) > 0 {
			f, ferr := fh[0].Open()
			if ferr == nil {
				// 10MB limit sanity wrap
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

		// Handle optional multiple media files
		files := r.MultipartForm.File["media"]

		// Validate media count before processing
		if len(files) > 4 {
			errors = append(errors, "Можно загрузить не более 4 изображений")
		} else {
			for _, fh := range files {
				f, ferr := fh.Open()
				if ferr != nil {
					h.Logger.ErrorContext(r.Context(), "error opening file", slog.Any("error", ferr))
					continue
				}
				// Storage handles streaming; service persists reference
				if _, uerr := h.LaunchService.AddMedia(r.Context(), l, fh.Filename, f); uerr != nil {
					h.Logger.ErrorContext(r.Context(), "error saving media", slog.Any("error", uerr))
					// Check if it's a media limit error
					if uerr.Error() == "too many media files" {
						errors = append(errors, "Можно загрузить не более 4 изображений")
						break // Stop processing more files
					}
				}
				f.Close()
			}
		}

		if len(errors) == 0 {
			p, _ := h.ProductService.GetByID(r.Context(), productID)
			slug := productID.String()
			if p != nil && p.Slug != "" {
				slug = p.Slug
			}
			redirectTo := "/products/" + slug + "/launches/edit"
			w.Header().Add("HX-Redirect", redirectTo)
			return
		}

	case launch.ErrLaunchAlreadyExists:
		errors = append(errors, "Запуск с таким названием уже существует")
		redirectTo := "/products/" + productID.String() + "/launches/edit"
		w.Header().Add("HX-Redirect", redirectTo)
		return

	case launch.ErrProductNotFound:
		errors = append(errors, "Продукт не найден")
		redirectTo := "/products"
		w.Header().Add("HX-Redirect", redirectTo)
		return

	case launch.ErrLaunchDateInPast:
		errors = append(errors, "Дата запуска не может быть в прошлом")
		redirectTo := "/products/" + productID.String() + "/launches/edit"
		w.Header().Add("HX-Redirect", redirectTo)
		return

	case launch.ErrTooManyMediaFiles:
		errors = append(errors, "Можно загрузить не более 4 изображений")
		redirectTo := "/products/" + productID.String() + "/launches/edit"
		w.Header().Add("HX-Redirect", redirectTo)
		return

	case launch.ErrInvalidURL:
		errors = append(errors, "Неверный URL")
		redirectTo := "/products/" + productID.String() + "/launches/edit"
		w.Header().Add("HX-Redirect", redirectTo)
		return
	default:
		h.Logger.ErrorContext(r.Context(), "error creating launch", slog.Any("error", err))
		errors = append(errors, "Что-то пошло не так. Пожалуйста, попробуйте еще раз.")
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
