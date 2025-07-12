package main

import (
	"RestApi/internal/config"
	"RestApi/internal/http-server/hadlers/redirect"
	detele_url "RestApi/internal/http-server/hadlers/url/detele-url"
	get_url "RestApi/internal/http-server/hadlers/url/get-url"
	"RestApi/internal/http-server/hadlers/url/save"
	mwLogger "RestApi/internal/http-server/middleware/logger"
	"RestApi/internal/lib/handlers/slogpretty"
	"RestApi/internal/storage/sqllite"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
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
		log.Error("failed to init storage", "error", err.Error())
		os.Exit(1)
	}
	log.Info("Starting db connection")

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))

		r.Post("/", save.New(log, storage))
		r.Post("/get-url", get_url.New(log, storage))
		r.Delete("/delete-url", detele_url.New(log, storage))
	})

	router.Get("/{alias}", redirect.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:              cfg.Address,
		Handler:           router,
		ReadHeaderTimeout: cfg.HTTPServer.Timeout,
		WriteTimeout:      cfg.HTTPServer.Timeout,
		IdleTimeout:       cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}
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
