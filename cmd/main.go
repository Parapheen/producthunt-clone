package main

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"

	"github.com/Parapheen/ph-clone/internal/app"
	"github.com/Parapheen/ph-clone/internal/infra/sqlite"
	"github.com/Parapheen/ph-clone/internal/infra/telegram"
	"github.com/Parapheen/ph-clone/internal/server"
	"github.com/Parapheen/ph-clone/internal/server/handler"
	mw "github.com/Parapheen/ph-clone/internal/server/middleware"
)

const (
	addr = ":3333"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	err := godotenv.Load(".env")
	if err != nil {
		logger.Error("Error loading .env file", "error", err)
		os.Exit(1)
	}

	db, err := sqlite.InitDB(os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Error("Error initializing database", "error", err)
		os.Exit(1)
	}

	telegramClient := telegram.NewTelegramClient(os.Getenv("TELEGRAM_BOT_TOKEN"), logger)

	userRepository := sqlite.NewUserRepository(db)
	productRepository := sqlite.NewProductRepository(db)
	launchRepository := sqlite.NewLaunchRepository(db)

	authService := app.NewAuthService(userRepository)
	userService := app.NewUserService(userRepository)
	productService := app.NewProductService(productRepository)
	launchService := app.NewLaunchService(launchRepository, telegramClient)

	m := mw.NewMiddleware(userService)

	h := handler.NewHandler(
		logger,
		authService,
		userService,
		productService,
		launchService,
	)
	s := server.NewServer(h, m)

	logger.Info("Staring server", "address", addr)
	s.Run(addr)
}
