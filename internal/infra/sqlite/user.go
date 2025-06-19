package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserModel struct {
	ID    uuid.UUID `db:"id"`
	Email string    `db:"email"`
	Name  string    `db:"name"`

	SessionID        *uuid.UUID `db:"session_id"`
	SessionToken     *string    `db:"session_token"`
	SessionExpiresAt *time.Time `db:"session_expires_at"`

	SocialAccountID         uuid.UUID `db:"social_account_id"`
	SocialAccountProvider   string    `db:"social_account_provider"`
	SocialAccountProviderID string    `db:"social_account_provider_id"`
}

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

func (r *UserRepository) GetByProvider(ctx context.Context, provider, providerID string) (*user.User, error) {
	query := `SELECT 
		u.id, u.email, u.name, 
		ss.id as session_id, ss.token as session_token, ss.expires_at as session_expires_at,
		s.id as social_account_id, s.provider as social_account_provider, s.provider_id as social_account_provider_id
		FROM users u
		JOIN social_accounts s ON u.id = s.user_id
		LEFT JOIN sessions ss ON u.id = ss.user_id
		WHERE s.provider = $1 AND s.provider_id = $2`
	uData := &UserModel{}

	err := r.db.GetContext(ctx, uData, query, provider, providerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	u := &user.User{
		ID:    uData.ID,
		Email: uData.Email,
		Name:  uData.Name,

		SocialAccounts: []*user.SocialAccount{
			{
				ID:         uData.SocialAccountID,
				Provider:   uData.SocialAccountProvider,
				ProviderID: uData.SocialAccountProviderID,
			},
		},
	}

	if uData.SessionID != nil {
		u.Session = &user.Session{
			ID:        *uData.SessionID,
			Token:     *uData.SessionToken,
			ExpiresAt: *uData.SessionExpiresAt,
		}
	}

	return u, nil
}

func (r *UserRepository) CreateSession(ctx context.Context, user *user.User) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO sessions (id, token, user_id, expires_at)
		VALUES ($1, $2, $3, $4)`,
		user.Session.ID,
		user.Session.Token,
		user.ID,
		user.Session.ExpiresAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) RefreshSession(ctx context.Context, session *user.Session) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE sessions 
		SET token = $1, expires_at = $2 WHERE id = $3`,
		session.Token,
		session.ExpiresAt,
		session.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) DeleteSession(ctx context.Context, session string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE token = ?`, session)
	if err != nil {
		return err
	}

	return nil
}
