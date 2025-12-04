// Package config provides configuration management using Viper
// Viper reads configuration from .env files
// See: https://github.com/spf13/viper#reading-config-files
package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// TODO: What happens for omit empty??
// TODO: Handle prod vs dev
// What about logging here??

type Config struct {
	AppEnv                   string   `mapstructure:"APP_ENV" validate:"omitempty"`
	Port                     int      `mapstructure:"PORT" validate:"min=1,max=65535"`
	PostgresConnectionString string   `mapstructure:"POSTGRES_CONNECTION_STRING" validate:"required"`
	ClickhouseURL            string   `mapstructure:"CLICKHOUSE_URL" validate:"required"`
	ClickhouseUsername       string   `mapstructure:"CLICKHOUSE_USERNAME" validate:"required"`
	ClickhousePassword       string   `mapstructure:"CLICKHOUSE_PASSWORD" validate:"required"`
	RedisURL                 string   `mapstructure:"REDIS_URL" validate:"required"`
	RedisUsername            string   `mapstructure:"REDIS_USERNAME" validate:"required"`
	RedisPassword            string   `mapstructure:"REDIS_PASSWORD" validate:"required"`
	RedisDB                  int      `mapstructure:"REDIS_DB" validate:"omitempty"`
	RedisDialTimeout         int      `mapstructure:"REDIS_DIAL_TIMEOUT" validate:"omitempty"`
	RedisReadTimeout         int      `mapstructure:"REDIS_READ_TIMEOUT" validate:"omitempty"`
	RedisWriteTimeout        int      `mapstructure:"REDIS_WRITE_TIMEOUT" validate:"omitempty"`
	RedisMaxRetries          int      `mapstructure:"REDIS_MAX_RETRIES" validate:"omitempty,min=1"`
	ClerkSecretKey           string   `mapstructure:"CLERK_SECRET_KEY" validate:"required"`
	CORSAllowedOrigins       []string `mapstructure:"CORS_ALLOWED_ORIGINS" validate:"omitempty"`
	CORSAllowedMethods       []string `mapstructure:"CORS_ALLOWED_METHODS" validate:"omitempty"`
	CORSAllowedHeaders       []string `mapstructure:"CORS_ALLOWED_HEADERS" validate:"omitempty"`
	CORSExposedHeaders       []string `mapstructure:"CORS_EXPOSED_HEADERS" validate:"omitempty"`
	CORSAllowCredentials     bool     `mapstructure:"CORS_ALLOW_CREDENTIALS" validate:"omitempty"`
	CORSMaxAge               int      `mapstructure:"CORS_MAX_AGE" validate:"omitempty"`
	ServerReadTimeout        int      `mapstructure:"SERVER_READ_TIMEOUT" validate:"min=1"`
	ServerWriteTimeout       int      `mapstructure:"SERVER_WRITE_TIMEOUT" validate:"min=1"`
	ServerIdleTimeout        int      `mapstructure:"SERVER_IDLE_TIMEOUT" validate:"min=1"`
}

var cfg *Config
var validate = validator.New()

// validateConfig validates the configuration using struct tags
func validateConfig(c *Config) error {
	if err := validate.Struct(c); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return fmt.Errorf("config validation failed: %w", err)
		}

		var errorMessages []string
		for _, fieldErr := range validationErrors {
			errorMessages = append(
				errorMessages,
				fmt.Sprintf("%s %s", fieldErr.Field(), getValidationErrorMessage(fieldErr)))
		}

		return fmt.Errorf("%s", strings.Join(errorMessages, "; "))
	}

	return nil
}

// getValidationErrorMessage returns a user-friendly error message for validation errors
func getValidationErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "is required"
	case "min":
		return fmt.Sprintf("must be at least %s", err.Param())
	case "max":
		return fmt.Sprintf("must be at most %s", err.Param())
	default:
		return err.Error()
	}
}

// parseCommaSeparated splits a comma-separated string into a slice, trimming whitespace
func parseCommaSeparated(s string) []string {
	if s == "" {
		return []string{}
	}

	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// Load reads configuration using Viper
func Load() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	v := viper.New()

	v.SetDefault("APP_ENV", "development")
	v.SetDefault("PORT", 8080)

	v.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000")
	v.SetDefault("CORS_ALLOWED_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
	v.SetDefault("CORS_ALLOWED_HEADERS", "Accept,Authorization,Content-Type,X-CSRF-Token")
	v.SetDefault("CORS_EXPOSED_HEADERS", "Link")
	v.SetDefault("CORS_ALLOW_CREDENTIALS", true)
	v.SetDefault("CORS_MAX_AGE", 300)
	v.SetDefault("SERVER_READ_TIMEOUT", 15)
	v.SetDefault("SERVER_WRITE_TIMEOUT", 15)
	v.SetDefault("SERVER_IDLE_TIMEOUT", 60)

	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("REDIS_DIAL_TIMEOUT", 5)
	v.SetDefault("REDIS_READ_TIMEOUT", 3)
	v.SetDefault("REDIS_WRITE_TIMEOUT", 3)
	v.SetDefault("REDIS_MAX_RETRIES", 3)

	// If running in a container use v.AutomaticEnv() to get platform's env vars
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath("./")

	cfg = &Config{}

	if err := v.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf(".env file not found: %w", err)
	}

	if err := v.Unmarshal(cfg); err != nil {
		return cfg, fmt.Errorf("Failed to unmarshal config: %w", err)
	}

	cfg.CORSAllowedOrigins = parseCommaSeparated(v.GetString("CORS_ALLOWED_ORIGINS"))
	cfg.CORSAllowedMethods = parseCommaSeparated(v.GetString("CORS_ALLOWED_METHODS"))
	cfg.CORSAllowedHeaders = parseCommaSeparated(v.GetString("CORS_ALLOWED_HEADERS"))
	cfg.CORSExposedHeaders = parseCommaSeparated(v.GetString("CORS_EXPOSED_HEADERS"))

	if err := validateConfig(cfg); err != nil {
		return cfg, fmt.Errorf("Config validation failed: %w", err)
	}

	return cfg, nil
}
