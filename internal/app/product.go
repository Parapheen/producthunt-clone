package app

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Parapheen/ph-clone/internal/domain/product"
	"github.com/google/uuid"
)

type ProductService struct {
	productRepo  product.ProductRepository
	categoryRepo product.CategoryRepository
    storage      Storage
    mailer       Mailer
    baseURL      string
}

func NewProductService(
	productRepo product.ProductRepository,
	categoryRepo product.CategoryRepository,
) *ProductService {
	return &ProductService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

func (s *ProductService) Create(ctx context.Context, p *product.Product) error {
	err := p.Validate()
	if err != nil {
		return err
	}

	err = s.productRepo.Create(ctx, p)
	if err != nil {
		return err
	}

	return nil
}

func (s *ProductService) NameExists(ctx context.Context, name string) (bool, error) {
	return s.productRepo.ExistsByName(ctx, name)
}

func (s *ProductService) URLExists(ctx context.Context, u string) (bool, error) {
	return s.productRepo.ExistsByURL(ctx, u)
}

func (s *ProductService) GetBySlug(ctx context.Context, slug string) (*product.Product, error) {
	return s.productRepo.GetBySlug(ctx, slug)
}

func (s *ProductService) GetByOwner(ctx context.Context, owner uuid.UUID) ([]*product.Product, error) {
	return s.productRepo.GetByOwner(ctx, owner)
}

func (s *ProductService) GetByMember(ctx context.Context, userID uuid.UUID) ([]*product.Product, error) {
    return s.productRepo.GetByMember(ctx, userID)
}

func (s *ProductService) GetByID(ctx context.Context, id uuid.UUID) (*product.Product, error) {
	return s.productRepo.GetByID(ctx, id)
}

func (s *ProductService) GetCategoryBySlug(ctx context.Context, slug string) (*product.Category, error) {
	return s.categoryRepo.GetBySlug(ctx, slug)
}

func (s *ProductService) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*product.Product, error) {
    return s.productRepo.GetByIDs(ctx, ids)
}

func (s *ProductService) GetByCategorySlug(ctx context.Context, slug string) ([]*product.Product, error) {
    return s.productRepo.GetByCategorySlug(ctx, slug)
}

func (s *ProductService) ListCategories(ctx context.Context) ([]*product.Category, error) {
    return s.categoryRepo.ListAll(ctx)
}

// WithStorage wires a storage dependency to the service (for DI without breaking existing constructors).
func (s *ProductService) WithStorage(storage Storage) *ProductService {
    s.storage = storage
    return s
}

// WithMailer wires a Mailer dependency for sending emails.
func (s *ProductService) WithMailer(mailer Mailer) *ProductService {
    s.mailer = mailer
    return s
}

// WithBaseURL wires the application's public base URL for building absolute links in emails.
func (s *ProductService) WithBaseURL(baseURL string) *ProductService {
    s.baseURL = baseURL
    return s
}

// UpdateImage uploads and sets a product image URL.
func (s *ProductService) UpdateImage(ctx context.Context, productID uuid.UUID, originalFilename string, content io.Reader) (string, error) {
    if s.storage == nil {
        return "", fmt.Errorf("storage not configured")
    }
    url, err := s.storage.Save(ctx, fmt.Sprintf("products/%s", productID.String()), originalFilename, content)
    if err != nil {
        return "", err
    }
    if err := s.productRepo.UpdateImageURL(ctx, productID, url); err != nil {
        return "", err
    }
    return url, nil
}

// UpdateTagline updates product tagline.
func (s *ProductService) UpdateTagline(ctx context.Context, productID uuid.UUID, tagline string) error {
    return s.productRepo.UpdateTagline(ctx, productID, tagline)
}

// InviteMember creates an invitation; if the email belongs to an existing user, we still send invite flow, then on accept add as member.
func (s *ProductService) InviteMember(ctx context.Context, productID uuid.UUID, email string, role product.Role) (*product.Invitation, error) {
    inv := &product.Invitation{
        ID:        uuid.New(),
        ProductID: productID,
        Email:     email,
        Role:      role,
        Token:     uuid.NewString(),
        Status:    product.InvitePending,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    if err := s.productRepo.CreateInvitation(ctx, inv); err != nil {
        return nil, err
    }
    // Send email if mailer is configured
    if s.mailer != nil {
        // Compose accept link
        // We fetch product to include its name in email
        p, _ := s.productRepo.GetByID(ctx, productID)
        acceptPath := "/invitations/accept?token=" + inv.Token
        acceptURL := acceptPath
        if s.baseURL != "" {
            acceptURL = s.baseURL + acceptPath
        }
        data := map[string]any{
            "ProductName": func() string { if p != nil { return p.Name }; return "Продукт" }(),
            "AcceptURL":   acceptURL,
        }
        _ = s.mailer.Send(ctx, email, "invite_member.html", data)
    }
    return inv, nil
}

// AcceptInvitation adds user as member by token and marks invitation accepted.
func (s *ProductService) AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) (uuid.UUID, error) {
    inv, err := s.productRepo.GetInvitationByToken(ctx, token)
    if err != nil {
        return uuid.Nil, err
    }
    if err := s.productRepo.AddMember(ctx, inv.ProductID, userID, inv.Role); err != nil {
        return uuid.Nil, err
    }
    if err := s.productRepo.MarkInvitationAccepted(ctx, token); err != nil {
        return uuid.Nil, err
    }
    return inv.ProductID, nil
}
