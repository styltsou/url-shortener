package config

import "testing"

func TestLoadFromEnvironmentWithoutDotEnv(t *testing.T) {
	cfg = nil
	t.Cleanup(func() { cfg = nil })
	t.Setenv("POSTGRES_CONNECTION_STRING", "postgres://example/test")
	t.Setenv("REDIS_URL", "redis:6379")
	t.Setenv("REDIS_USERNAME", "cache-user")
	t.Setenv("CLERK_SECRET_KEY", "sk_test_dummy")
	t.Setenv("CLICKHOUSE_URL", "clickhouse:9000")

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.PostgresConnectionString != "postgres://example/test" || loaded.RedisURL != "redis:6379" ||
		loaded.RedisUsername != "cache-user" || loaded.ClerkSecretKey != "sk_test_dummy" ||
		loaded.ClickhouseURL != "clickhouse:9000" {
		t.Fatal("environment variables were not loaded into the configuration")
	}
}
