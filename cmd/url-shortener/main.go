package main

import (
	"RestApi/internal/config"
	"RestApi/internal/http-server/handlers/redirect"
	"RestApi/internal/http-server/handlers/url/delete"
	"RestApi/internal/http-server/handlers/url/get"
	"RestApi/internal/http-server/handlers/url/save"
	mwLogger "RestApi/internal/http-server/middleware/logger"
	"RestApi/internal/lib/handlers/slogpretty"
	"RestApi/internal/storage/postgres"
	"RestApi/storage/scripts"
	"flag"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
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
	cfg := initializeConfig()
	fmt.Println("cfg__________->", cfg.GetDBURL())

	if shouldRunMigrations() {
		runMigrations(cfg)
		return
	}

	logger := setupLogger(cfg.Env)
	logStartupInfo(logger, cfg.Env)

	storage := initializeStorage(logger, cfg)
	router := setupRouter(logger, cfg, storage)

	startServer(logger, cfg, router)
}

func initializeConfig() *config.Config {
	flag.Parse()
	return config.MustLoad()
}

func shouldRunMigrations() bool {
	migrateFlag := flag.Bool("migrate", false, "Run database migration")
	return *migrateFlag
}

func runMigrations(cfg *config.Config) {
	if err := scripts.RunMigrations(cfg.GetDBURL()); err != nil {
		log.Fatalf("Migrations failed: %v", err)
	}
	log.Println("Migrations completed successfully")
}

func setupLogger(env string) *slog.Logger {
	switch env {
	case envLocal:
		return setupPrettySlog()
	case envDev:
		return slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{Level: slog.LevelDebug}))
	default: // prod
		return slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{Level: slog.LevelInfo}))
	}
}

func logStartupInfo(logger *slog.Logger, env string) {
	logger.Info("Starting RestApi application", slog.String("env", env))
	logger.Debug("Debug messages are enabled")
}

func initializeStorage(logger *slog.Logger, cfg *config.Config) *postgres.Storage {
	storage, err := postgres.New(cfg.GetDBURL(), cfg.HTTPServer.Timeout)
	if err != nil {
		logger.Error("Failed to initialize storage", "error", err.Error())
		os.Exit(1)
	}
	logger.Info("Database connection established")
	return storage
}

func setupRouter(logger *slog.Logger, cfg *config.Config, storage *postgres.Storage) *chi.Mux {
	router := chi.NewRouter()

	// Common middleware
	router.Use(
		middleware.RequestID,
		middleware.Logger,
		mwLogger.New(logger),
		middleware.Recoverer,
		middleware.URLFormat,
	)

	// Protected routes
	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))

		r.Post("/", save.New(logger, storage))
		r.Post("/get-url", get.New(logger, storage))
		r.Delete("/delete-url", delete.New(logger, storage))
	})

	// Public route
	router.Get("/{alias}", redirect.New(logger, storage))

	return router
}

func startServer(logger *slog.Logger, cfg *config.Config, router *chi.Mux) {
	server := &http.Server{
		Addr:              cfg.HTTPServer.Address,
		Handler:           router,
		ReadHeaderTimeout: cfg.HTTPServer.Timeout,
		WriteTimeout:      cfg.HTTPServer.Timeout,
		IdleTimeout:       cfg.HTTPServer.IdleTimeout,
	}

	logger.Info("Starting server", slog.String("address", server.Addr))
	if err := server.ListenAndServe(); err != nil {
		logger.Error("Failed to start server", "error", err)
	}
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}
	return slog.New(opts.NewPrettyHandler(os.Stdout))
}
