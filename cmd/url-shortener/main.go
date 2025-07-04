package main

import (
	"RestApi/internal/config"
	mwLogger "RestApi/internal/http-server/middleware/logger"
	"RestApi/internal/lib/handlers/slogpretty"
	"RestApi/internal/storage/sqllite"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
		log.Error(fmt.Sprintf("for alias \"%s\" failed to retrieve url", alias), err.Error())
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

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	//TODO: run server
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
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

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
