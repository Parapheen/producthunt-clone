package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sort"

	"github.com/Parapheen/ph-clone/internal/domain/launch"
	"github.com/google/uuid"
)

type LaunchService struct {
	launchRepo     launch.LaunchRepository
	telegramCleint TelegramClient
}

func NewLaunchService(
	launchRepo launch.LaunchRepository,
	telegramCleint TelegramClient,
) *LaunchService {
	return &LaunchService{
		launchRepo:     launchRepo,
		telegramCleint: telegramCleint,
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

	if launch.InModeration() {
		go func() {
			msg := fmt.Sprintf(
				`justlaunch 🚀

Продукт %s отправил новый запуск (%s) на модерацию

#justlaunch`,
				launch.ProductID,
				launch.ID,
			)

			err := s.telegramCleint.Send(
				context.WithoutCancel(ctx),
				os.Getenv("TELEGRAM_CHAT_ID"),
				msg,
			)

			if err != nil {
				slog.Error("error sending message to admin", slog.Any("err", err))
			}
		}()
	}

	return s.launchRepo.Update(ctx, launch)
}

func (s *LaunchService) GetByOwner(ctx context.Context, ownerID uuid.UUID) ([]*launch.Launch, error) {
	return s.launchRepo.GetByOwner(ctx, ownerID)
}

func (s *LaunchService) GetPublishedByProduct(ctx context.Context, productID uuid.UUID) ([]*launch.Launch, error) {
	launches, err := s.launchRepo.GetByProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	publishedLaunches := make([]*launch.Launch, 0, len(launches))
	for _, l := range launches {
		if l.State == launch.Published && l.LaunchDate != nil {
			publishedLaunches = append(publishedLaunches, l)
		}
	}
	sort.Slice(publishedLaunches, func(i, j int) bool {
		return publishedLaunches[i].LaunchDate.After(*publishedLaunches[j].LaunchDate)
	})

	return publishedLaunches, nil
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

func (s *LaunchService) GetByState(ctx context.Context, states []launch.State) ([]*launch.Launch, error) {
	return s.launchRepo.GetByState(ctx, states)
}
