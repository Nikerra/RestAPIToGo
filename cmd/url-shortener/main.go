package main

import (
	"RestApi/internal/config"
	"RestApi/internal/storage/sqllite"
	"fmt"
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	cfg := config.MustLoad()
	fmt.Println("File configuration to read.")

	log := setupLogger(cfg.Env)
	log.Info("Starting RestApi application",
		slog.String("env", cfg.Env))
	log.Debug("Debug messages are enabled")

	storage, err := sqllite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", err.Error())
		os.Exit(1)
	}

	_ = storage
	log.Info("Starting db connection")

	//TODO: router: chi, "chi render"

	//TODO: run server
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
