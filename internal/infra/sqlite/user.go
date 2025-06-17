package sqlite

import (
	"context"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	return runInTx(ctx, r.db, func(tx *sqlx.Tx) error {
		_, err := tx.NamedExec(`
			INSERT INTO users (id, email, name)
			VALUES (:id, :email, :name)
		`, u)
		if err != nil {
			return err
		}

		for _, socialAccount := range u.SocialAccounts {
			_, err := tx.ExecContext(ctx, `
				INSERT INTO social_accounts (id, provider, provider_id, user_id)
				VALUES (:id, :provider, :provider_id, :user_id)
			`, socialAccount.ID, socialAccount.Provider, socialAccount.ProviderID, u.ID)
			if err != nil {
				return err
			}
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO sessions (id, token, user_id, expires_at)
			VALUES (:id, :token, :user_id, :expires_at)
		`, u.Session.ID, u.Session.Token, u.ID, u.Session.ExpiresAt)
		if err != nil {
			return err
		}

		return nil
	})
}

func (r *UserRepository) GetBySession(ctx context.Context, session string) (*user.User, error) {
	query := `SELECT u.id, u.email, u.name FROM users u
		INNER JOIN sessions s ON u.id = s.user_id
		WHERE s.token = ?`
	u := &user.User{}

	err := r.db.GetContext(ctx, u, query, session)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (r *UserRepository) DeleteSession(ctx context.Context, session string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE token = ?`, session)
	if err != nil {
		return err
	}

	return nil
}
