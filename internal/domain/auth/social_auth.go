package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

func (s *Service) Authenticate(ctx context.Context, code string) (*YandexUser, error) {
	token, err := s.YandexConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("error exchanging code: %w", err)
	}

	resp, err := s.YandexConfig.Client(ctx, token).Get("https://login.yandex.ru/info")
	if err != nil {
		return nil, fmt.Errorf("error getting user info: %w", err)
	}

	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	var yandexUser YandexUser
	err = json.Unmarshal(body, &yandexUser)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling user info: %w", err)
	}

	return &yandexUser, nil
}

func (s *Service) GetRedirectURL(state string) string {
	return s.YandexConfig.AuthCodeURL(state)
}
