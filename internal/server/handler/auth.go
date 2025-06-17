package handler

import (
	"html/template"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (h *Handler) LoginModal(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("views/partials/auth-modal.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) YandexAuth(w http.ResponseWriter, r *http.Request) {
	state := uuid.NewString()
	http.SetCookie(w, &http.Cookie{
		Name:     "ouath_state",
		Value:    state,
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Now().Add(time.Minute * 10),
	})

	url := h.AuthService.GetSocialRedirectURL("yandex", state)

	http.Redirect(w, r, url, http.StatusFound)
}

func (h *Handler) YandexAuthCallback(w http.ResponseWriter, r *http.Request) {
	// validate state
	cookie, err := r.Cookie("ouath_state")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	state := r.URL.Query().Get("state")
	if cookie.Value != state {
		http.Error(w, "Invalid state", http.StatusInternalServerError)
		return
	}

	code := r.URL.Query().Get("code")

	user, err := h.AuthService.AuthenticateWithSocial(r.Context(), "yandex", code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// clean state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "ouath_state",
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Unix(0, 0),
	})

	// set session cookie
	sessionCookie := &http.Cookie{
		Name:     "session",
		Value:    user.Session.Token,
		HttpOnly: true,
		Path:     "/",
		Expires:  user.Session.ExpiresAt,
	}
	http.SetCookie(w, sessionCookie)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.AuthService.Logout(r.Context(), cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Unix(0, 0),
	})

	http.Redirect(w, r, "/", http.StatusFound)
}
