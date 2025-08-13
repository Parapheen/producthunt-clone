package handler

import (
	"context"
	"io"

	"github.com/Parapheen/ph-clone/internal/app"
	"github.com/Parapheen/ph-clone/internal/domain/launch"
	"github.com/Parapheen/ph-clone/internal/domain/product"
	"github.com/Parapheen/ph-clone/internal/domain/user"
	"github.com/google/uuid"
)

type AuthService interface {
	GetSocialRedirectURL(provider, state string) string
	AuthenticateWithSocial(ctx context.Context, provider, code string) (*user.User, error)
	Logout(ctx context.Context, session string) error
}

type UserService interface {
	GetBySession(ctx context.Context, session string) (*user.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*user.User, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*user.User, error)
    UpdateAvatar(ctx context.Context, userID uuid.UUID, originalFilename string, content io.Reader) (string, error)
    UpdateBio(ctx context.Context, userID uuid.UUID, bio string) error
}

type ProductService interface {
	Create(ctx context.Context, product *product.Product) error
	NameExists(ctx context.Context, name string) (bool, error)
	URLExists(ctx context.Context, u string) (bool, error)
	GetBySlug(ctx context.Context, slug string) (*product.Product, error)
	GetByOwner(ctx context.Context, owner uuid.UUID) ([]*product.Product, error)
    GetByMember(ctx context.Context, userID uuid.UUID) ([]*product.Product, error)
	GetByID(ctx context.Context, id uuid.UUID) (*product.Product, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*product.Category, error)
    GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*product.Product, error)
    UpdateImage(ctx context.Context, productID uuid.UUID, originalFilename string, content io.Reader) (string, error)
    UpdateTagline(ctx context.Context, productID uuid.UUID, tagline string) error
    InviteMember(ctx context.Context, productID uuid.UUID, email string, role product.Role) (*product.Invitation, error)
    AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) (uuid.UUID, error)
    // Catalog
    GetByCategorySlug(ctx context.Context, slug string) ([]*product.Product, error)
    ListCategories(ctx context.Context) ([]*product.Category, error)
}

type LaunchService interface {
	GetBySlug(ctx context.Context, slug string) (*launch.Launch, error)
    GetByProductAndSlug(ctx context.Context, productID uuid.UUID, slug string) (*launch.Launch, error)
	GetLatestByProduct(ctx context.Context, productID uuid.UUID) (*launch.Launch, error)
	Update(ctx context.Context, launch *launch.Launch) error
	GetByOwner(ctx context.Context, ownerID uuid.UUID) ([]*launch.Launch, error)
	GetByProduct(ctx context.Context, productID uuid.UUID) ([]*launch.Launch, error)
	GetPublishedByProduct(ctx context.Context, productID uuid.UUID) ([]*launch.Launch, error)
	GetByID(ctx context.Context, id uuid.UUID) (*launch.Launch, error)
	GetFeed(ctx context.Context) ([]*launch.Launch, error)
    GetFeedByPeriod(ctx context.Context, period string) ([]*launch.Launch, error)
	GetByState(ctx context.Context, states []launch.State) ([]*launch.Launch, error)
	Create(ctx context.Context, launch *launch.Launch) error
	Delete(ctx context.Context, launchID uuid.UUID) error
    AddMedia(ctx context.Context, l *launch.Launch, originalFilename string, content io.Reader) (string, error)
    ReplaceMedia(ctx context.Context, l *launch.Launch, files []app.FileUpload) error
    UpdateImage(ctx context.Context, launchID uuid.UUID, originalFilename string, content io.Reader) (string, error)
    ToggleUpvote(ctx context.Context, launchID, userID uuid.UUID) (bool, int, error)

    // Comments
    CreateComment(ctx context.Context, c *launch.Comment) error
    GetCommentsTree(ctx context.Context, launchID uuid.UUID) ([]*launch.Comment, map[uuid.UUID][]*launch.Comment, error)
    ToggleCommentUpvote(ctx context.Context, commentID, userID uuid.UUID) (bool, int, error)
    PinComment(ctx context.Context, commentID uuid.UUID, pinned bool) error

    // Admin notifications
    SendAdminNotification(ctx context.Context, message string) error
}

type Storage interface {
    Save(ctx context.Context, pathPrefix string, originalFilename string, r io.Reader) (publicURL string, err error)
    Delete(ctx context.Context, publicURL string) error
}
