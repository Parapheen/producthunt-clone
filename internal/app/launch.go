package app

import (
	"context"

	"github.com/Parapheen/ph-clone/internal/domain/launch"
	"github.com/google/uuid"
)

type LaunchService struct {
	launchRepo launch.LaunchRepository
}

func NewLaunchService(launchRepo launch.LaunchRepository) *LaunchService {
	return &LaunchService{
		launchRepo: launchRepo,
	}
}

func (s *LaunchService) GetBySlug(ctx context.Context, slug string) (*launch.Launch, error) {
	return s.launchRepo.GetBySlug(ctx, slug)
}

func (s *LaunchService) GetByID(ctx context.Context, id uuid.UUID) (*launch.Launch, error) {
	return s.launchRepo.GetByID(ctx, id)
}

func (s *LaunchService) GetLatestByProduct(ctx context.Context, productID uuid.UUID) (*launch.Launch, error) {
	return s.launchRepo.GetLatestByProduct(ctx, productID)
}

func (s *LaunchService) Update(ctx context.Context, launch *launch.Launch) error {
	err := launch.Validate()
	if err != nil {
		return err
	}

	return s.launchRepo.Update(ctx, launch)
}

func (s *LaunchService) GetByOwner(ctx context.Context, ownerID uuid.UUID) ([]*launch.Launch, error) {
	return s.launchRepo.GetByOwner(ctx, ownerID)
}

func (s *LaunchService) GetByProduct(ctx context.Context, productID uuid.UUID) ([]*launch.Launch, error) {
	return s.launchRepo.GetByProduct(ctx, productID)
}

func (s *LaunchService) GetFeed(ctx context.Context) ([]*launch.Launch, error) {
	return s.launchRepo.GetFeed(ctx, "all_time", 100, 0)
}

func (s *LaunchService) Delete(ctx context.Context, launchID uuid.UUID) error {
	return s.launchRepo.Delete(ctx, launchID)
}

func (s *LaunchService) Create(ctx context.Context, launch *launch.Launch) error {
	return s.launchRepo.Create(ctx, launch)
}
