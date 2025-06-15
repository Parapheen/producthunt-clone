package main

import (
	"log/slog"
	"os"

	"github.com/Parapheen/ph-clone/internal/domain/auth"
	"github.com/Parapheen/ph-clone/internal/infra/sqlite"
	"github.com/Parapheen/ph-clone/internal/server"
	"github.com/Parapheen/ph-clone/internal/server/handler"
	"github.com/joho/godotenv"
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

	authService := auth.NewService()

	h := handler.NewHandler(authService)
	s := server.NewServer(h)

	logger.Info("Staring server", "address", addr)
	s.Run(addr)
}
