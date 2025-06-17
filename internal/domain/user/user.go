package user

import "github.com/google/uuid"

type User struct {
	ID    uuid.UUID
	Email string
	Name  string

	Session        *Session
	SocialAccounts []*SocialAccount
}

func NewUserFromSocialAccount(socialAccount *SocialAccount) *User {
	return &User{
		ID:    uuid.New(),
		Email: socialAccount.Email,
		Name:  socialAccount.Name,

		Session:        NewSession(),
		SocialAccounts: []*SocialAccount{socialAccount},
	}
}
