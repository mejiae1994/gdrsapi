package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	CloudflareApiKey    string
	CloudflareAccountId string
	GeminiApiKey        string
	GoogleSACred        string
	Environment         string
}

var (
	cfg  *Config
	once sync.Once
)

func GetConfig() (*Config, error) {
	var err error
	once.Do(func() {
		err = LoadConfig()
	})

	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func LoadConfig() error {
	loadedCfg := false
	// Load env from docker compose
	cfId := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	cfApiKey := os.Getenv("CLOUDFLARE_API_KEY")
	geminiApiKey := os.Getenv("GEMINI_API_KEY")
	googleSACred := os.Getenv("GOOGLE_SA_CRED")

	loadedCfg = !(cfId == "" || cfApiKey == "" || geminiApiKey == "" || googleSACred == "")

	if loadedCfg {
		cfg = &Config{
			CloudflareAccountId: cfId,
			CloudflareApiKey:    cfApiKey,
			GeminiApiKey:        geminiApiKey,
			GoogleSACred:        googleSACred,
		}
		return nil
	}

	// Load env from .env file in Dev
	err := godotenv.Load(".env")
	if err != nil {
		// this is used for testing
		err = godotenv.Load("../../.env")
		if err != nil {
			return fmt.Errorf("Error loading .env file: %w", err)
		}
	}

	cfg = &Config{
		CloudflareAccountId: os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
		CloudflareApiKey:    os.Getenv("CLOUDFLARE_API_KEY"),
		GeminiApiKey:        os.Getenv("GEMINI_API_KEY"),
		GoogleSACred:        os.Getenv("GOOGLE_SA_CRED"),
		Environment:         os.Getenv("ENVIRONMENT"),
	}

	// Required env vars
	if cfg.CloudflareAccountId == "" {
		return fmt.Errorf("Cloudflare account id is not set")
	}

	if cfg.CloudflareApiKey == "" {
		return fmt.Errorf("Cloudflare api key is not set")
	}

	if cfg.GeminiApiKey == "" {
		return fmt.Errorf("Gemini api key is not set")
	}

	if cfg.GoogleSACred == "" {
		return fmt.Errorf("Google service account credentials are not set")
	}

	return nil
}
