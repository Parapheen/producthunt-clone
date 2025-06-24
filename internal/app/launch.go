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

func (s *LaunchService) GetLatestByProduct(ctx context.Context, productID uuid.UUID) (*launch.Launch, error) {
	return s.launchRepo.GetLatestByProduct(ctx, productID)
}

func (s *LaunchService) Update(ctx context.Context, launch *launch.Launch) error {
	return s.launchRepo.Update(ctx, launch)
}
