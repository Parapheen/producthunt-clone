package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/goforj/godump"
	"github.com/google/uuid"
)

// VKOauthProvider implements VK ID (id.vk.com) OAuth without SDK using PKCE
// Docs: https://id.vk.com/about/business/go/docs/ru/vkid/latest/vk-id/connection/start-integration/auth-without-sdk/auth-without-sdk-web

type VKOauthProvider struct {
	clientID    string
	redirectURL string
	scope       string
	httpClient  *http.Client
}

func NewVKOauthProvider() *VKOauthProvider {
	scope := os.Getenv("VK_SCOPE")
	if strings.TrimSpace(scope) == "" {
		// Request email VK ID defaults also include vkid.personal_info
		scope = "email"
	}
	return &VKOauthProvider{
		clientID:    os.Getenv("VK_CLIENT_ID"),
		redirectURL: os.Getenv("VK_REDIRECT_URL"),
		scope:       scope,
		httpClient:  http.DefaultClient,
	}
}

// BuildAuthURL constructs the VK ID authorization URL including PKCE challenge.
func (v *VKOauthProvider) BuildAuthURL(state string, codeChallenge string) string {
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", v.clientID)
	if v.scope != "" {
		q.Set("scope", v.scope)
	}
	q.Set("redirect_uri", v.redirectURL)
	q.Set("state", state)
	if codeChallenge != "" {
		q.Set("code_challenge", codeChallenge)
		q.Set("code_challenge_method", "S256")
	}
	return "https://id.vk.com/authorize?" + q.Encode()
}

// vkTokenResponse matches VK ID token exchange response

type vkTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	ExpiresIn    int    `json:"expires_in"`
	UserID       int64  `json:"user_id"`
	State        string `json:"state"`
	Scope        string `json:"scope"`
}

// ExchangeCode exchanges authorization code for tokens using PKCE verifier and device_id.
func (v *VKOauthProvider) ExchangeCode(ctx context.Context, code string, codeVerifier string, deviceID string, state string) (*vkTokenResponse, error) {
	form := url.Values{}
	form.Set("client_id", v.clientID)
	form.Set("grant_type", "authorization_code")
	form.Set("code_verifier", codeVerifier)
	form.Set("device_id", deviceID)
	form.Set("code", code)
	form.Set("redirect_uri", v.redirectURL)
	if state != "" {
		form.Set("state", state)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://id.vk.com/oauth2/auth", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("vk: building token request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vk: token request failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("vk: token endpoint status %d: %s", resp.StatusCode, string(body))
	}

	var tok vkTokenResponse
	if err := json.Unmarshal(body, &tok); err != nil {
		return nil, fmt.Errorf("vk: token response decode failed: %w", err)
	}
	return &tok, nil
}

// RefreshAccessToken exchanges refresh_token for a new access token
func (v *VKOauthProvider) RefreshAccessToken(ctx context.Context, refreshToken string, deviceID string, state string) (*vkTokenResponse, error) {
	form := url.Values{}
	form.Set("client_id", v.clientID)
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("device_id", deviceID)
	if state != "" {
		form.Set("state", state)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://id.vk.com/oauth2/auth", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("vk: building refresh request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vk: refresh request failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("vk: refresh endpoint status %d: %s", resp.StatusCode, string(body))
	}
	var tok vkTokenResponse
	if err := json.Unmarshal(body, &tok); err != nil {
		return nil, fmt.Errorf("vk: refresh response decode failed: %w", err)
	}
	return &tok, nil
}

// vkUserInfoResponse matches VK ID user_info endpoint

type vkUserInfoResponse struct {
	User struct {
		UserID    string `json:"user_id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
		Avatar    string `json:"avatar"`
		Email     string `json:"email"`
	} `json:"user"`
}

// GetUserInfo retrieves user info using access_token
func (v *VKOauthProvider) GetUserInfo(ctx context.Context, accessToken string) (*user.SocialAccount, error) {
	form := url.Values{}
	form.Set("access_token", accessToken)
	form.Set("client_id", v.clientID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://id.vk.com/oauth2/user_info", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("vk: building user_info request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vk: user_info request failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("vk: user_info status %d: %s", resp.StatusCode, string(body))
	}

	var ui vkUserInfoResponse
	if err := json.Unmarshal(body, &ui); err != nil {
		return nil, fmt.Errorf("vk: user_info decode failed: %w", err)
	}

	godump.Dump(accessToken)
	godump.Dump(ui)

	email := strings.TrimSpace(ui.User.Email)
	if email == "" {
		return nil, errors.New("vk: email permission not granted or email not available")
	}

	name := strings.TrimSpace(strings.Trim(ui.User.FirstName+" "+ui.User.LastName, " "))
	return &user.SocialAccount{
		ID:         uuid.New(),
		Provider:   "vk",
		ProviderID: ui.User.UserID,
		Email:      email,
		Name:       name,
		AvatarURL:  ui.User.Avatar,
	}, nil
}
