package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const oauthStateCookie = "oauth_state"
const sessionCookieName = "session"
const inviteTokenCookie = "invite_token"
const vkPKCECookie = "vk_pkce_verifier"

func (h *Handler) LoginModal(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("views/partials/auth-modal.html")
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		h.InternalServerError(w, r, err)
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
		h.InternalServerError(w, r, err)
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
		h.InternalServerError(w, r, err)
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
		h.InternalServerError(w, r, err)
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
		h.InternalServerError(w, r, err)
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

	// PKCE for VK ID
	verifier := generatePKCEVerifier()
	challenge := pkceS256(verifier)
	http.SetCookie(w, &http.Cookie{
		Name:     vkPKCECookie,
		Value:    verifier,
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Now().Add(time.Minute * 10),
	})

	url := h.AuthService.GetVKAuthURL(state, challenge)

	http.Redirect(w, r, url, http.StatusFound)
}

func (h *Handler) VKAuthCallback(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(oauthStateCookie)
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Invalid code", http.StatusInternalServerError)
		return
	}
	deviceID := r.URL.Query().Get("device_id")
	if deviceID == "" {
		http.Error(w, "Invalid device_id", http.StatusInternalServerError)
		return
	}
	state := r.URL.Query().Get("state")
	if state == "" {
		http.Error(w, "Invalid state", http.StatusInternalServerError)
		return
	}

	if cookie.Value != state || state == "" {
		http.Error(w, "Invalid state", http.StatusInternalServerError)
		return
	}

	verifierCookie, err := r.Cookie(vkPKCECookie)
	if err != nil || strings.TrimSpace(verifierCookie.Value) == "" {
		http.Error(w, "Missing PKCE verifier", http.StatusInternalServerError)
		return
	}

	user, err := h.AuthService.AuthenticateWithVK(r.Context(), code, verifierCookie.Value, deviceID, state)
	if err != nil {
		h.InternalServerError(w, r, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Unix(0, 0),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     vkPKCECookie,
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
		h.InternalServerError(w, r, err)
		return
	}

	err = h.AuthService.Logout(r.Context(), cookie.Value)
	if err != nil {
		h.InternalServerError(w, r, err)
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

func generatePKCEVerifier() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return uuid.NewString()
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func pkceS256(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
