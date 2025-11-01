package config

import (
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

var (
	once sync.Once

	// Read-only globals (exported)
	AppEnv         string // "dev" | "prod" | etc.
	Port           int
	DatabaseURL    string
	ClerkSecretKey string
)

func Load() {
	once.Do(func() {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		AppEnv = getEnv("APP_ENV", "dev")
		Port = getEnvInt("PORT", 5000)
		DatabaseURL = mustGetEnv("DATABASE_URL")
		ClerkSecretKey = mustGetEnv("CLERK_SECRET_KEY")

		// basic validation
		if DatabaseURL == "" {
			log.Fatal("DATABASE_URL is required")
		}
	})
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func mustGetEnv(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("%s is required", key)
	}
	return v
}
