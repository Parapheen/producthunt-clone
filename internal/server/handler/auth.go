package handler

import (
	"html/template"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const oauthStateCookie = "oauth_state"
const sessionCookieName = "session"
const inviteTokenCookie = "invite_token"

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
		Name:     oauthStateCookie,
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
	cookie, err := r.Cookie(oauthStateCookie)
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
		Name:     oauthStateCookie,
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Unix(0, 0),
	})

	// set session cookie
	sessionCookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    user.Session.Token,
		HttpOnly: true,
		Path:     "/",
		Expires:  user.Session.ExpiresAt,
	}
	http.SetCookie(w, sessionCookie)

	if inv, err := r.Cookie(inviteTokenCookie); err == nil && inv.Value != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     inviteTokenCookie,
			Value:    "",
			HttpOnly: true,
			Path:     "/",
			Expires:  time.Unix(0, 0),
		})
		if productID, err := h.ProductService.AcceptInvitation(r.Context(), inv.Value, user.ID); err == nil {
			http.Redirect(w, r, "/products/u/"+productID.String(), http.StatusFound)
			return
		}
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handler) GoogleAuth(w http.ResponseWriter, r *http.Request) {
	state := uuid.NewString()
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    state,
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Now().Add(time.Minute * 10),
	})

	url := h.AuthService.GetSocialRedirectURL("google", state)

	http.Redirect(w, r, url, http.StatusFound)
}

func (h *Handler) GoogleAuthCallback(w http.ResponseWriter, r *http.Request) {
	// validate state
	cookie, err := r.Cookie(oauthStateCookie)
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

	user, err := h.AuthService.AuthenticateWithSocial(r.Context(), "google", code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// clean state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Unix(0, 0),
	})

	// set session cookie
	sessionCookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    user.Session.Token,
		HttpOnly: true,
		Path:     "/",
		Expires:  user.Session.ExpiresAt,
	}
	http.SetCookie(w, sessionCookie)

	if inv, err := r.Cookie(inviteTokenCookie); err == nil && inv.Value != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     inviteTokenCookie,
			Value:    "",
			HttpOnly: true,
			Path:     "/",
			Expires:  time.Unix(0, 0),
		})
		if productID, err := h.ProductService.AcceptInvitation(r.Context(), inv.Value, user.ID); err == nil {
			http.Redirect(w, r, "/products/u/"+productID.String(), http.StatusFound)
			return
		}
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handler) VKAuth(w http.ResponseWriter, r *http.Request) {
	state := uuid.NewString()
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    state,
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Now().Add(time.Minute * 10),
	})

	url := h.AuthService.GetSocialRedirectURL("vk", state)

	http.Redirect(w, r, url, http.StatusFound)
}

func (h *Handler) VKAuthCallback(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(oauthStateCookie)
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

	user, err := h.AuthService.AuthenticateWithSocial(r.Context(), "vk", code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Unix(0, 0),
	})

	sessionCookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    user.Session.Token,
		HttpOnly: true,
		Path:     "/",
		Expires:  user.Session.ExpiresAt,
	}
	http.SetCookie(w, sessionCookie)

	if inv, err := r.Cookie(inviteTokenCookie); err == nil && inv.Value != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     inviteTokenCookie,
			Value:    "",
			HttpOnly: true,
			Path:     "/",
			Expires:  time.Unix(0, 0),
		})
		if productID, err := h.ProductService.AcceptInvitation(r.Context(), inv.Value, user.ID); err == nil {
			http.Redirect(w, r, "/products/u/"+productID.String(), http.StatusFound)
			return
		}
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
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
		Name:     sessionCookieName,
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Unix(0, 0),
	})

	http.Redirect(w, r, "/", http.StatusFound)
}
