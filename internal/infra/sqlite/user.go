package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserModel struct {
	ID        uuid.UUID      `db:"id"`
	Email     string         `db:"email"`
	Name      string         `db:"name"`
	AvatarURL sql.NullString `db:"avatar_url"`
	Bio        sql.NullString `db:"bio"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`

	SessionID        *uuid.UUID `db:"session_id"`
	SessionToken     *string    `db:"session_token"`
	SessionExpiresAt *time.Time `db:"session_expires_at"`
}

type SocialAccountModel struct {
	ID         uuid.UUID `db:"id"`
	Provider   string    `db:"provider"`
	ProviderID string    `db:"provider_id"`
	UserID     uuid.UUID `db:"user_id"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	return runInTx(ctx, r.db, func(tx *sqlx.Tx) error {
        userInsert := struct {
            ID        uuid.UUID `db:"id"`
            Email     string    `db:"email"`
            Name      string    `db:"name"`
            AvatarURL string    `db:"avatar_url"`
        }{
            ID:        u.ID,
            Email:     u.Email,
            Name:      u.Name,
            AvatarURL: u.AvatarURL,
        }

        _, err := tx.NamedExecContext(ctx, `
            INSERT INTO users (id, email, name, avatar_url)
            VALUES (:id, :email, :name, :avatar_url)
        `, userInsert)
		if err != nil {
			return err
		}

		for _, socialAccount := range u.SocialAccounts {
			socialAccountInsert := SocialAccountModel{
				ID:         socialAccount.ID,
				Provider:   socialAccount.Provider,
				ProviderID: socialAccount.ProviderID,
				UserID:     u.ID,
			}
			_, err := tx.NamedExecContext(ctx, `
				INSERT INTO social_accounts (id, provider, provider_id, user_id)
				VALUES (:id, :provider, :provider_id, :user_id)
			`, socialAccountInsert)
			if err != nil {
				return err
			}
		}

		sessionInsert := struct {
			ID        uuid.UUID `db:"id"`
			Token     string    `db:"token"`
			UserID    uuid.UUID `db:"user_id"`
			ExpiresAt time.Time `db:"expires_at"`
		}{
			ID:        u.Session.ID,
			Token:     u.Session.Token,
			UserID:    u.ID,
			ExpiresAt: u.Session.ExpiresAt,
		}
		_, err = tx.NamedExecContext(ctx, `
			INSERT INTO sessions (id, token, user_id, expires_at)
			VALUES (:id, :token, :user_id, :expires_at)
		`, sessionInsert)
		if err != nil {
			return err
		}

		return nil
	})
}

func (r *UserRepository) GetBySession(ctx context.Context, sessionToken string) (*user.User, error) {
	query := `
		SELECT
            u.id, u.email, u.name, u.avatar_url, u.bio, u.created_at,
			s.id as session_id, s.token as session_token, s.expires_at as session_expires_at
		FROM users u
		INNER JOIN sessions s ON u.id = s.user_id
		WHERE s.token = $1`
	var uData UserModel

	if err := r.db.GetContext(ctx, &uData, query, sessionToken); err != nil {
		return nil, err
	}

	u := toDomainUser(&uData)

	return r.loadSocialAccounts(ctx, u)
}

func (r *UserRepository) GetByProvider(ctx context.Context, provider, providerID string) (*user.User, error) {
	query := `
		SELECT
            u.id, u.email, u.name, u.avatar_url, u.bio, u.created_at,
			ss.id as session_id, ss.token as session_token, ss.expires_at as session_expires_at
		FROM users u
		INNER JOIN social_accounts s ON u.id = s.user_id
		LEFT JOIN sessions ss ON u.id = ss.user_id
		WHERE s.provider = $1 AND s.provider_id = $2`
	var uData UserModel

	if err := r.db.GetContext(ctx, &uData, query, provider, providerID); err != nil {
		return nil, err
	}

	u := toDomainUser(&uData)

	return r.loadSocialAccounts(ctx, u)
}

func (r *UserRepository) CreateSession(ctx context.Context, user *user.User) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO sessions (id, token, user_id, expires_at) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET token = $2, expires_at = $4, updated_at = current_timestamp;
		`,
		user.Session.ID,
		user.Session.Token,
		user.ID,
		user.Session.ExpiresAt,
	)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	query := `
		SELECT
            u.id, u.email, u.name, u.avatar_url, u.bio, u.created_at,
			s.id as session_id, s.token as session_token, s.expires_at as session_expires_at
		FROM users u
		LEFT JOIN sessions s ON u.id = s.user_id
		WHERE u.id = $1`
	var uData UserModel

	if err := r.db.GetContext(ctx, &uData, query, id); err != nil {
		return nil, err
	}

	u := toDomainUser(&uData)

	return r.loadSocialAccounts(ctx, u)
}

func (r *UserRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*user.User, error) {
	query := `SELECT id, email, name, avatar_url, bio, created_at FROM users WHERE id IN (?)`

	query, args, err := sqlx.In(query, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to construct query for GetByIDs: %w", err)
	}

	// Rebind the query to use the correct placeholder format for SQLite
	query = r.db.Rebind(query)

	var users []*UserModel
	if err := r.db.SelectContext(ctx, &users, query, args...); err != nil {
		return nil, fmt.Errorf("failed to get users by IDs: %w", err)
	}

	domainUsers := make([]*user.User, 0, len(users))
	for _, u := range users {
		domainUsers = append(domainUsers, toDomainUser(u))
	}

	return domainUsers, nil
}

func (r *UserRepository) RefreshSession(ctx context.Context, session *user.Session) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE sessions SET token = $1, expires_at = $2, updated_at = current_timestamp WHERE id = $3`,
		session.Token,
		session.ExpiresAt,
		session.ID,
	)
	return err
}

func (r *UserRepository) UpdateAvatarURL(ctx context.Context, userID uuid.UUID, avatarURL string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET avatar_url = $1, updated_at = current_timestamp WHERE id = $2`, avatarURL, userID)
	return err
}

func (r *UserRepository) UpdateBio(ctx context.Context, userID uuid.UUID, bio string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET bio = $1, updated_at = current_timestamp WHERE id = $2`, bio, userID)
	return err
}

func (r *UserRepository) DeleteSession(ctx context.Context, sessionToken string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE token = $1`, sessionToken)
	return err
}

func (r *UserRepository) loadSocialAccounts(ctx context.Context, u *user.User) (*user.User, error) {
	if u == nil {
		return nil, nil
	}

	var socialAccountModels []SocialAccountModel
	query := `SELECT * FROM social_accounts WHERE user_id = $1`

	err := r.db.SelectContext(ctx, &socialAccountModels, query, u.ID)
	if err != nil {
		return nil, err
	}

	u.SocialAccounts = make([]*user.SocialAccount, 0, len(socialAccountModels))
	for _, saModel := range socialAccountModels {
		u.SocialAccounts = append(u.SocialAccounts, &user.SocialAccount{
			ID:         saModel.ID,
			Provider:   saModel.Provider,
			ProviderID: saModel.ProviderID,
		})
	}

	return u, nil
}

func toDomainUser(uData *UserModel) *user.User {
	if uData == nil {
		return nil
	}

	u := &user.User{
		ID:        uData.ID,
		Email:     uData.Email,
		Name:      uData.Name,
		AvatarURL: uData.AvatarURL.String,
		Bio:       uData.Bio.String,
		CreatedAt: uData.CreatedAt,
	}

	if uData.SessionID != nil {
		u.Session = &user.Session{
			ID:        *uData.SessionID,
			Token:     *uData.SessionToken,
			ExpiresAt: *uData.SessionExpiresAt,
		}
	}

	return u
}
