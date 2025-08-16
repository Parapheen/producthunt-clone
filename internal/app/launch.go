package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sort"
	"time"

	"github.com/Parapheen/ph-clone/internal/domain/launch"
	"github.com/google/uuid"
)

type LaunchService struct {
	launchRepo     launch.LaunchRepository
	telegramCleint TelegramClient
	storage        Storage
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

// GetByProductAndSlug retrieves a launch by product and slug
func (s *LaunchService) GetByProductAndSlug(ctx context.Context, productID uuid.UUID, slug string) (*launch.Launch, error) {
	return s.launchRepo.GetByProductAndSlug(ctx, productID, slug)
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

// GetUpvotedMap returns a map of launchID->bool indicating whether the user has upvoted each launch ID
func (s *LaunchService) GetUpvotedMap(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (map[uuid.UUID]bool, error) {
	return s.launchRepo.GetUpvotedByUserForLaunchIDs(ctx, userID, ids)
}

func (s *LaunchService) GetFeed(ctx context.Context) ([]*launch.Launch, error) {
	return s.launchRepo.GetFeed(ctx, "all_time", 100, 0)
}

// GetFeedByPeriod returns the feed filtered by the given period.
// Valid periods: "daily", "weekly", "monthly", "all_time". Defaults to "daily".
func (s *LaunchService) GetFeedByPeriod(ctx context.Context, period string) ([]*launch.Launch, error) {
	normalized := normalizePeriod(period)
	return s.launchRepo.GetFeed(ctx, normalized, 100, 0)
}

// GetNthByProductOrderedByCreatedAt returns the nth published launch of the product by created_at DESC.
func (s *LaunchService) GetNthByProductOrderedByCreatedAt(ctx context.Context, productID uuid.UUID, index int) (*launch.Launch, error) {
	if index <= 0 {
		return nil, fmt.Errorf("invalid index")
	}
	return s.launchRepo.GetNthByProductOrderedByCreatedAt(ctx, productID, index)
}

// GetIndexByProductAndLaunchID returns the 1-based index by created_at DESC for a launch within a product.
func (s *LaunchService) GetIndexByProductAndLaunchID(ctx context.Context, productID, launchID uuid.UUID) (int, error) {
	return s.launchRepo.GetIndexByProductAndLaunchID(ctx, productID, launchID)
}

func normalizePeriod(period string) string {
	switch period {
	case "daily", "weekly", "monthly", "all_time":
		return period
	default:
		return "daily"
	}
}

func (s *LaunchService) Delete(ctx context.Context, launchID uuid.UUID) error {
	return s.launchRepo.Delete(ctx, launchID)
}

func (s *LaunchService) Create(ctx context.Context, l *launch.Launch) error {
	err := l.Validate()
	if err != nil {
		return err
	}

	return s.launchRepo.Create(ctx, l)
}

func (s *LaunchService) GetByState(ctx context.Context, states []launch.State) ([]*launch.Launch, error) {
	return s.launchRepo.GetByState(ctx, states)
}

func (s *LaunchService) WithStorage(storage Storage) *LaunchService {
	s.storage = storage
	return s
}

// AddMedia uploads a launch media image and appends URL to Launch.Media.
func (s *LaunchService) AddMedia(ctx context.Context, l *launch.Launch, originalFilename string, content io.Reader) (string, error) {
	if s.storage == nil {
		return "", fmt.Errorf("storage not configured")
	}

	// Check if adding this media would exceed the limit
	if len(l.Media) >= 4 {
		return "", launch.ErrTooManyMediaFiles
	}

	url, err := s.storage.Save(ctx, fmt.Sprintf("launches/%s", l.ID.String()), originalFilename, content)
	if err != nil {
		return "", err
	}
	// Persist media reference in repository and update in-memory entity
	if err := s.launchRepo.AddMedia(ctx, l.ID, url); err != nil {
		return "", err
	}
	l.Media = append(l.Media, url)

	// Validate after adding
	if err := l.Validate(); err != nil {
		return "", err
	}

	return url, nil
}

// UpdateImage uploads a launch avatar image and persists its URL
func (s *LaunchService) UpdateImage(ctx context.Context, launchID uuid.UUID, originalFilename string, content io.Reader) (string, error) {
	if s.storage == nil {
		return "", fmt.Errorf("storage not configured")
	}
	url, err := s.storage.Save(ctx, fmt.Sprintf("launches/%s/avatar", launchID.String()), originalFilename, content)
	if err != nil {
		return "", err
	}
	if err := s.launchRepo.UpdateImageURL(ctx, launchID, url); err != nil {
		return "", err
	}
	return url, nil
}

// ReplaceMedia replaces all existing media with new uploaded files.
func (s *LaunchService) ReplaceMedia(ctx context.Context, l *launch.Launch, files []FileUpload) error {
	if s.storage == nil {
		return fmt.Errorf("storage not configured")
	}

	// Validate media count before processing
	if len(files) > 4 {
		return launch.ErrTooManyMediaFiles
	}

	// Delete old media from storage
	for _, oldURL := range l.Media {
		if err := s.storage.Delete(ctx, oldURL); err != nil {
			// Log error but continue - we don't want to fail the whole operation
			slog.Error("failed to delete old media", slog.String("url", oldURL), slog.Any("error", err))
		}
	}

	// Upload new files
	newURLs := make([]string, 0, len(files))
	for _, file := range files {
		url, err := s.storage.Save(ctx, fmt.Sprintf("launches/%s", l.ID.String()), file.Filename, file.Content)
		if err != nil {
			return fmt.Errorf("failed to upload %s: %w", file.Filename, err)
		}
		newURLs = append(newURLs, url)
	}

	// Update database
	if err := s.launchRepo.ReplaceMedia(ctx, l.ID, newURLs); err != nil {
		return err
	}

	// Update in-memory entity and validate
	l.Media = newURLs
	return l.Validate()
}

// FileUpload represents a file to be uploaded
type FileUpload struct {
	Filename string
	Content  io.Reader
}

// ToggleUpvote toggles the upvote and returns the new upvoted state and total count.
func (s *LaunchService) ToggleUpvote(ctx context.Context, launchID, userID uuid.UUID) (bool, int, error) {
	return s.launchRepo.ToggleUpvote(ctx, launchID, userID)
}

// --- Comments ---

func (s *LaunchService) CreateComment(ctx context.Context, c *launch.Comment) error {
	if err := c.Validate(); err != nil {
		return err
	}
	return s.launchRepo.CreateComment(ctx, c)
}

func (s *LaunchService) GetCommentsTree(ctx context.Context, launchID uuid.UUID) ([]*launch.Comment, map[uuid.UUID][]*launch.Comment, error) {
	return s.launchRepo.GetCommentsTree(ctx, launchID)
}

func (s *LaunchService) ToggleCommentUpvote(ctx context.Context, commentID, userID uuid.UUID) (bool, int, error) {
	return s.launchRepo.ToggleCommentUpvote(ctx, commentID, userID)
}

func (s *LaunchService) PinComment(ctx context.Context, commentID uuid.UUID, pinned bool) error {
	return s.launchRepo.PinComment(ctx, commentID, pinned)
}

// SendAdminNotification sends a free-form admin notification via Telegram.
func (s *LaunchService) SendAdminNotification(ctx context.Context, message string) error {
	return s.telegramCleint.Send(ctx, os.Getenv("TELEGRAM_CHAT_ID"), message)
}

// --- Awards ---

func (s *LaunchService) ListAwards(ctx context.Context) ([]*launch.Award, error) {
	return s.launchRepo.ListAwards(ctx)
}

func (s *LaunchService) AssignAwardToLaunch(ctx context.Context, launchID uuid.UUID, awardCode string, periodDate time.Time) error {
	return s.launchRepo.AssignAward(ctx, launchID, awardCode, periodDate)
}

func (s *LaunchService) GetAwardsByLaunchIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID][]*launch.LaunchAward, error) {
	return s.launchRepo.GetAwardsByLaunchIDs(ctx, ids)
}

// HasAwardForPeriod returns whether the award already exists for the period
func (s *LaunchService) HasAwardForPeriod(ctx context.Context, awardCode string, periodDate time.Time) (bool, error) {
	return s.launchRepo.HasAwardForPeriod(ctx, awardCode, periodDate)
}

// GetTopLaunchInRange returns the top launch by upvotes in [start, end)
func (s *LaunchService) GetTopLaunchInRange(ctx context.Context, start, end time.Time) (*launch.Launch, error) {
	return s.launchRepo.GetTopLaunchInRange(ctx, start, end)
}
