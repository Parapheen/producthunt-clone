package main

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"

	"github.com/Parapheen/ph-clone/internal/app"
	"github.com/Parapheen/ph-clone/internal/infra/mailer"
	"github.com/Parapheen/ph-clone/internal/infra/sqlite"
	localstorage "github.com/Parapheen/ph-clone/internal/infra/storage/local"
	"github.com/Parapheen/ph-clone/internal/infra/telegram"
	"github.com/Parapheen/ph-clone/internal/pkg/config"
	"github.com/Parapheen/ph-clone/internal/server"
	"github.com/Parapheen/ph-clone/internal/server/handler"
	mw "github.com/Parapheen/ph-clone/internal/server/middleware"
)

func main() {
	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		logger.Warn("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize database
	db, err := sqlite.InitDB(cfg.Database.URL)
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize Telegram client
	telegramClient := telegram.NewTelegramClient(cfg.Telegram.BotToken, logger)

	// Initialize repositories
	userRepository := sqlite.NewUserRepository(db)
	productRepository := sqlite.NewProductRepository(db)
	launchRepository := sqlite.NewLaunchRepository(db)
	categoryRepository := sqlite.NewCategoryRepository(db)

	// Initialize services
	authService := app.NewAuthService(userRepository)
    userService := app.NewUserService(userRepository)
    productService := app.NewProductService(productRepository, categoryRepository)
    launchService := app.NewLaunchService(launchRepository, telegramClient)

    // Initialize storage (local by default)
    var storage app.Storage
    switch cfg.Storage.Driver {
    case "local":
        storage = localstorage.NewFilesystemStorage(cfg.Storage.LocalUploadDir, cfg.Storage.PublicUploadBase)
    default:
        storage = localstorage.NewFilesystemStorage(cfg.Storage.LocalUploadDir, cfg.Storage.PublicUploadBase)
    }

    // Wire storage into services
    _ = userService.WithStorage(storage)
    _ = productService.WithStorage(storage)
    _ = launchService.WithStorage(storage)

    // Initialize mailer (dummy or smtp depending on env)
    // Mailer: use SMTP in prod when configured, otherwise dummy
    if cfg.IsProduction() && cfg.SMTP.Host != "" {
        _ = productService.WithMailer(mailer.NewSMTPMailer(logger, cfg.SMTP))
    } else {
        _ = productService.WithMailer(mailer.NewDummyMailer(logger))
    }
    _ = productService.WithBaseURL(cfg.App.BaseURL)

	// Initialize middleware
	m := mw.NewMiddleware(userService)

	// Initialize handler
	h := handler.NewHandler(
		logger,
		db,
		authService,
		userService,
		productService,
		launchService,
        storage,
	)

	// Initialize server
	s := server.NewServer(h, m, cfg)

	// Start server
	logger.Info("Starting server", 
		"port", cfg.Server.Port,
		"environment", cfg.App.Environment,
		"base_url", cfg.App.BaseURL,
	)

	if err := s.Run(); err != nil {
		logger.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
