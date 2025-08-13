package product

import (
	"context"

	"github.com/google/uuid"
)

type ProductRepository interface {
	Create(ctx context.Context, product *Product) error
	ExistsByName(ctx context.Context, name string) (bool, error)
	ExistsByURL(ctx context.Context, url string) (bool, error)
	GetBySlug(ctx context.Context, slug string) (*Product, error)
	GetByOwner(ctx context.Context, owner uuid.UUID) ([]*Product, error)
    GetByMember(ctx context.Context, userID uuid.UUID) ([]*Product, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
    GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*Product, error)
    UpdateImageURL(ctx context.Context, productID uuid.UUID, imageURL string) error
    UpdateTagline(ctx context.Context, productID uuid.UUID, tagline string) error

    // Catalog
    GetByCategorySlug(ctx context.Context, slug string) ([]*Product, error)

    // Invitations
    CreateInvitation(ctx context.Context, inv *Invitation) error
    GetInvitationByToken(ctx context.Context, token string) (*Invitation, error)
    MarkInvitationAccepted(ctx context.Context, token string) error
    RevokeInvitation(ctx context.Context, token string) error
    AddMember(ctx context.Context, productID, userID uuid.UUID, role Role) error
}

type CategoryRepository interface {
	GetBySlug(ctx context.Context, slug string) (*Category, error)
    ListAll(ctx context.Context) ([]*Category, error)
}
