package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/Parapheen/ph-clone/internal/pkg/tmpl"
	"github.com/justinas/nosurf"
)

// PromotingPage renders the advertising/promoting information page with an example
func (h *Handler) PromotingPage(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())

	t, err := template.New("promoting").Funcs(template.FuncMap{
		"dict": tmpl.Dict,
	}).ParseFiles(
		"views/promoting.html",
		"views/layout/layout.html",
		"views/layout/header.html",
		"views/layout/footer.html",
		"views/layout/head.html",
		"views/partials/promotion-result.html",
	)
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}

	meta := map[string]any{
		"Title":       "Продвижение запуска — justlaunch",
		"Description": "Реклама запуска: закреплённая карточка между другими запусками в ленте и категориях.",
		"Canonical":   h.BaseURL + "/promoting",
	}

	if err := t.ExecuteTemplate(w, "layout", map[string]any{
		"User":  u,
		"token": nosurf.Token(r),
		"meta":  meta,
	}); err != nil {
		h.InternalServerError(w, r, err)
		return
	}
}

// RequestPromotion handles CTA: sends a Telegram message with logged-in user's email
func (h *Handler) RequestPromotion(w http.ResponseWriter, r *http.Request) {
	u := user.GetUserFromContext(r.Context())
	if u == nil {
		w.WriteHeader(http.StatusUnauthorized)
		t, err := template.ParseFiles("views/partials/promotion-result.html")
		if err != nil {
			h.InternalServerError(w, r, err)
			return
		}
		_ = t.ExecuteTemplate(w, "promotion-result", map[string]any{
			"Title":        "Требуется вход",
			"Message":      "Пожалуйста, войдите, чтобы отправить заявку на продвижение",
			"RequireLogin": true,
			"Support":      true,
			"ShowBack":     false,
		})
		return
	}

	// Compose admin notification
	msg := fmt.Sprintf(
		`justlaunch 🚀
Запрос на продвижение
Email: %s
UserID: %s
When: %s

#promote`,
		u.Email,
		u.ID,
		time.Now().Format(time.RFC3339),
	)

	if err := h.LaunchService.SendAdminNotification(r.Context(), msg); err != nil {
		h.Logger.ErrorContext(r.Context(), "error sending promotion request", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		t, tplErr := template.ParseFiles("views/partials/promotion-result.html")
		if tplErr != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		_ = t.ExecuteTemplate(w, "promotion-result", map[string]any{
			"Title":    "Ошибка",
			"Message":  "Не удалось отправить запрос. Попробуйте позже",
			"Support":  true,
			"ShowBack": false,
		})
		return
	}

	t, err := template.ParseFiles("views/partials/promotion-result.html")
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}
	_ = t.ExecuteTemplate(w, "promotion-result", map[string]any{
		"Title":    "Заявка отправлена",
		"Message":  "Спасибо! Мы свяжемся с вами в течение 24 часов",
		"Support":  true,
		"ShowBack": false,
	})
}
