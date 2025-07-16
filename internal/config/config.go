package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	Env         string `yaml:"env" env:"ENV" env-default:"local"`
	StoragePath string `yaml:"storage_path" env:"STORAGE_PATH"`
	Database    struct {
		Host    string `yaml:"host" env:"DB_HOST"`
		Port    string `yaml:"port" env:"DB_PORT"`
		Name    string `yaml:"name" env:"DB_NAME"`
		SSLMode string `yaml:"ssl_mode" env:"DB_SSLMODE"`
	} `yaml:"database"`
	HTTPServer `yaml:"http_server"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env:"HTTP_ADDRESS"`
	Timeout     time.Duration `yaml:"timeout" env:"HTTP_TIMEOUT"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env:"HTTP_IDLE_TIMEOUT"`
	User        string        `yaml:"user" env:"HTTP_USER"`
	Password    string        `yaml:"password" env:"HTTP_PASSWORD"`
}

func MustLoad() *Config {
	// load .env standard storage
	loadEnvFiles()

	// load config
	var cfg Config
	if err := cleanenv.ReadConfig("config/local.yaml", &cfg); err != nil {
		log.Fatalf("Cannot read config: %v", err)
	}

	// check require vars
	checkRequiredEnvVars()

	return &cfg
}

func loadEnvFiles() {
	envPaths := []string{
		filepath.Join("env", ".env"),
		".env",
		"/app/env/.env",
	}

	for _, path := range envPaths {
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Load(path); err == nil {
				log.Printf("Loaded env from %s", path)
				return
			}
		}
	}
	log.Printf("Warning: no .env file found in %v", envPaths)
}

func checkRequiredEnvVars() {
	required := []string{
		"DB_USER",
		"DB_PASS",
		"HTTP_USER",
		"HTTP_PASSWORD",
	}

	for _, varName := range required {
		if os.Getenv(varName) == "" {
			log.Fatalf("Required environment variable %s is not set", varName)
		}
	}
}

func (c *Config) GetDBURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		c.Database.Host,
		os.Getenv("DB_PORT_IN"),
		c.Database.Name,
		c.Database.SSLMode,
	)
}
