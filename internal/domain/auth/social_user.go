package auth

type YandexUser struct {
	ID    string `json:"id"`
	Name  string `json:"display_name"`
	Email string `json:"default_email"`
}
