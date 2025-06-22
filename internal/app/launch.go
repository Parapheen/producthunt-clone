package app

import (
	"context"

	"github.com/Parapheen/ph-clone/internal/domain/launch"
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
