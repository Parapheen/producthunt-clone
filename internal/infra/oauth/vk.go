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
	"strconv"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// VK OAuth2 endpoints are not included in x/oauth2 providers by default, so we define them.
var vkEndpoint = oauth2.Endpoint{
    AuthURL:  "https://oauth.vk.com/authorize",
    TokenURL: "https://oauth.vk.com/access_token",
}

type VKOauthProvider struct {
    config *oauth2.Config
    apiVer string
}

func NewVKOauthProvider() *VKOauthProvider {
    return &VKOauthProvider{
        config: &oauth2.Config{
            ClientID:     os.Getenv("VK_CLIENT_ID"),
            ClientSecret: os.Getenv("VK_CLIENT_SECRET"),
            Endpoint:     vkEndpoint,
            RedirectURL:  os.Getenv("VK_REDIRECT_URL"),
            Scopes:       []string{"email"},
        },
        apiVer: "5.131",
    }
}

func (v *VKOauthProvider) GetAuthCodeURL(state string) string {
    // VK supports adding display or scope params; oauth2.Config already includes scope
    return v.config.AuthCodeURL(state)
}

func (v *VKOauthProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
    token, err := v.config.Exchange(ctx, code)
    if err != nil {
        return nil, fmt.Errorf("vk: error exchanging code: %w", err)
    }
    return token, nil
}

type vkUsersGetResponse struct {
    Response []struct {
        ID        int    `json:"id"`
        FirstName string `json:"first_name"`
        LastName  string `json:"last_name"`
        Photo200  string `json:"photo_200"`
    } `json:"response"`
    Error *struct {
        ErrorCode int    `json:"error_code"`
        ErrorMsg  string `json:"error_msg"`
    } `json:"error"`
}

func (v *VKOauthProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*user.SocialAccount, error) {
    if !token.Valid() {
        return nil, errors.New("vk: invalid token")
    }

    // Try to read email from token extras (VK returns email in token response when scope includes email)
    email := ""
    if e := token.Extra("email"); e != nil {
        email = fmt.Sprint(e)
    }

    // Also capture user_id from token extras if provided
    providerID := ""
    if uid := token.Extra("user_id"); uid != nil {
        switch u := uid.(type) {
        case string:
            providerID = u
        case json.Number:
            providerID = u.String()
        case float64:
            providerID = strconv.FormatInt(int64(u), 10)
        case int64:
            providerID = strconv.FormatInt(u, 10)
        case int:
            providerID = strconv.Itoa(u)
        default:
            providerID = fmt.Sprint(uid)
        }
    }

    // Call VK API to get profile info (name, etc.)
    // Even if providerID is empty, users.get will return it from the API
    values := url.Values{}
    values.Set("access_token", token.AccessToken)
    values.Set("v", v.apiVer)
    values.Set("fields", "photo_200")

    apiURL := "https://api.vk.com/method/users.get?" + values.Encode()
    req, _ := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("vk: error calling users.get: %w", err)
    }
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)

    var vkr vkUsersGetResponse
    if err := json.Unmarshal(body, &vkr); err != nil {
        return nil, fmt.Errorf("vk: error decoding users.get response: %w", err)
    }
    if vkr.Error != nil {
        return nil, fmt.Errorf("vk: api error %d: %s", vkr.Error.ErrorCode, vkr.Error.ErrorMsg)
    }
    if len(vkr.Response) == 0 {
        return nil, errors.New("vk: empty users.get response")
    }

    u := vkr.Response[0]
    if providerID == "" {
        providerID = strconv.Itoa(u.ID)
    }
    name := fmt.Sprintf("%s %s", u.FirstName, u.LastName)

    if email == "" {
        return nil, errors.New("vk: email permission not granted or email not available")
    }

    return &user.SocialAccount{
        ID:         uuid.New(),
        Provider:   "vk",
        ProviderID: providerID,
        Email:      email,
        Name:       name,
        AvatarURL:  u.Photo200,
    }, nil
}


