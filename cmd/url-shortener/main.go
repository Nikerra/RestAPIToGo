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
	log.Info("Starting db connection")
	///////////////////////////////////////////////////////////////////////
	//id, err := storage.SaveURL("https://google.com", "google")
	//if err != nil {
	//	log.Error("failed to save url", err.Error())
	//	os.Exit(1)
	//}
	//log.Info("Successfully saved url", slog.Int64("id", id))

	alias := "google"
	resURL, err := storage.GetURL(alias)
	if err != nil {
		log.Error("failed to retrieve url", err.Error())
	} else {
		log.Info(fmt.Sprintf("Get url for alias=%s, url=%s", alias, resURL))
	}

	alias = "yandex"
	resURL, err = storage.GetURL(alias)
	if err != nil {
		log.Error("failed to retrieve url", err.Error())
	} else {
		log.Info(fmt.Sprintf("Get url for alias=%s, url=%s", alias, resURL))
	}

	alias = "google"
	err = storage.DeleteURL(alias)
	if err != nil {
		log.Error("failed to delete url", err.Error())
	} else {
		log.Info(fmt.Sprintf("Delete url for alias=%s", alias))
	}

	id, err := storage.SaveURL("https://google.com", "google")
	if err != nil {
		log.Error("failed to save url", err.Error())
		os.Exit(1)
	}
	log.Info("Successfully saved url", slog.Int64("id", id))
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
