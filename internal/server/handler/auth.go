package handler

import (
	"fmt"
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

	url := h.AuthService.GetRedirectURL(state)

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

	user, err := h.AuthService.Authenticate(r.Context(), code)

	// user service login or create and save session cookie
	fmt.Printf("%+v\n", user)

	// clean state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "ouath_state",
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Unix(0, 0),
	})
	http.Redirect(w, r, "/", http.StatusFound)
}
