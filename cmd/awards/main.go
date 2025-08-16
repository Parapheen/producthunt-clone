package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/Parapheen/ph-clone/internal/app"
	"github.com/Parapheen/ph-clone/internal/domain/launch"
	"github.com/Parapheen/ph-clone/internal/infra/sqlite"
	"github.com/Parapheen/ph-clone/internal/pkg/config"
)

// normalizeWeek returns the Monday (UTC) of the ISO week for a given date
func normalizeWeek(t time.Time) time.Time {
	wd := int(t.Weekday())
	if wd == 0 {
		wd = 7
	}
	delta := wd - 1
	return time.Date(t.Year(), t.Month(), t.Day()-delta, 0, 0, 0, 0, time.UTC)
}

func main() {
	// Flags: allow overriding the reference date (defaults to now)
	refStr := flag.String("date", "", "Reference date in YYYY-MM-DD; awards will be computed for the last completed day/week/year relative to this date")
	dryRun := flag.Bool("dry", false, "Print actions without writing to DB")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	_ = godotenv.Load(".env")

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config load", "err", err)
		os.Exit(1)
	}

	db, err := sqlite.InitDB(cfg.Database.URL)
	if err != nil {
		logger.Error("db init", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	launchRepo := sqlite.NewLaunchRepository(db)
	launchSvc := app.NewLaunchService(launchRepo, nil)

	ctx := context.Background()

	// Determine reference time (UTC)
	now := time.Now().UTC()
	if *refStr != "" {
		parsed, perr := time.Parse("2006-01-02", *refStr)
		if perr != nil {
			logger.Error("invalid -date", "value", *refStr, "err", perr)
			os.Exit(1)
		}
		now = parsed
	}

	// Compute last completed periods
	yesterday := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC)
	lastWeekMonday := normalizeWeek(now).AddDate(0, 0, -7)
	lastMonthFirst := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastMonthEndExclusive := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	lastYearFirst := time.Date(now.Year()-1, 1, 1, 0, 0, 0, 0, time.UTC)
	lastYearEndExclusive := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)

	assign := func(code string, periodDate time.Time, start, end time.Time) error {
		has, err := launchRepo.HasAwardForPeriod(ctx, code, periodDate)
		if err != nil {
			return err
		}
		if has {
			logger.Info("award exists, skipping", "award", code, "period", periodDate.Format("2006-01-02"))
			return nil
		}
		top, err := launchRepo.GetTopLaunchInRange(ctx, start, end)
		if err != nil {
			if err.Error() == "sql: no rows in result set" {
				logger.Info("no launches in range", "award", code, "start", start, "end", end)
				return nil
			}
			return err
		}
		if *dryRun {
			fmt.Printf("DRY: would assign %s to launch %s for period %s\n", code, top.ID.String(), periodDate.Format("2006-01-02"))
			return nil
		}
		if err := launchSvc.AssignAwardToLaunch(ctx, top.ID, code, periodDate); err != nil {
			return err
		}
		logger.Info("assigned award", "award", code, "launch", top.ID.String(), "period", periodDate.Format("2006-01-02"))
		return nil
	}

	// Product of the day: use full previous UTC day
	if err := assign(launch.AwardCodeProductOfDay, yesterday, yesterday, yesterday.AddDate(0, 0, 1)); err != nil {
		logger.Error("assign daily", "err", err)
	}

	// Product of the week: ISO week starting previous Monday
	if err := assign(launch.AwardCodeProductOfWeek, lastWeekMonday, lastWeekMonday, lastWeekMonday.AddDate(0, 0, 7)); err != nil {
		logger.Error("assign weekly", "err", err)
	}

	// Product of the month: last full calendar month
	if err := assign(launch.AwardCodeProductOfMonth, lastMonthFirst, lastMonthFirst, lastMonthEndExclusive); err != nil {
		logger.Error("assign monthly", "err", err)
	}

	// Product of the year: last full calendar year
	if err := assign(launch.AwardCodeProductOfYear, lastYearFirst, lastYearFirst, lastYearEndExclusive); err != nil {
		logger.Error("assign yearly", "err", err)
	}
}
