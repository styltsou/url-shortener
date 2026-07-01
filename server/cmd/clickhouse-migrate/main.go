package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/spf13/viper"
)

const (
	migrationsDir = "migrations/clickhouse"
)

type migration struct {
	version string
	name    string
	upSQL   string
	downSQL string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: clickhouse-migrate [up|down|status]")
		os.Exit(1)
	}

	command := os.Args[1]

	v := viper.New()
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath("./")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Error reading .env file: %v\n", err)
			os.Exit(1)
		}
	}

	url := v.GetString("CLICKHOUSE_URL")
	if url == "" {
		fmt.Println("CLICKHOUSE_URL is not set")
		os.Exit(1)
	}

	username := v.GetString("CLICKHOUSE_USERNAME")
	password := v.GetString("CLICKHOUSE_PASSWORD")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{url},
		Auth: clickhouse.Auth{
			Username: username,
			Password: password,
		},
		DialTimeout: 10 * time.Second,
	})
	if err != nil {
		fmt.Printf("Failed to open ClickHouse connection: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	if err := conn.Ping(ctx); err != nil {
		fmt.Printf("Failed to ping ClickHouse: %v\n", err)
		os.Exit(1)
	}

	if err := ensureMigrationsTable(ctx, conn); err != nil {
		fmt.Printf("Failed to create migrations table: %v\n", err)
		os.Exit(1)
	}

	migrations, err := loadMigrations()
	if err != nil {
		fmt.Printf("Failed to load migrations: %v\n", err)
		os.Exit(1)
	}

	if len(migrations) == 0 {
		fmt.Println("No migration files found")
		return
	}

	switch command {
	case "up":
		runUp(ctx, conn, migrations)
	case "down":
		runDown(ctx, conn, migrations)
	case "status":
		runStatus(ctx, conn, migrations)
	default:
		fmt.Printf("Unknown command: %s (use: up, down, status)\n", command)
		os.Exit(1)
	}
}

func ensureMigrationsTable(ctx context.Context, conn driver.Conn) error {
	return conn.Exec(ctx, `
		CREATE DATABASE IF NOT EXISTS link4it;
		CREATE TABLE IF NOT EXISTS link4it.schema_migrations (
			version     String,
			applied_at  DateTime DEFAULT now()
		) ENGINE = MergeTree()
		ORDER BY version
	`)
}

func loadMigrations() ([]migration, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory %s: %w", migrationsDir, err)
	}

	filesByVersion := make(map[string]struct {
		upFile   string
		downFile string
		name     string
	})

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		parts := strings.SplitN(name, "_", 2)
		if len(parts) < 2 {
			continue
		}
		version := parts[0]

		suffix := ".up.sql"
		isUp := strings.HasSuffix(name, ".up.sql")
		if !isUp {
			suffix = ".down.sql"
			if !strings.HasSuffix(name, ".down.sql") {
				continue
			}
		}

		baseName := strings.TrimSuffix(name, suffix)
		nameParts := strings.SplitN(baseName, "_", 2)
		desc := ""
		if len(nameParts) >= 2 {
			desc = nameParts[1]
		}

		entry := filesByVersion[version]
		entry.name = desc
		if isUp {
			entry.upFile = filepath.Join(migrationsDir, name)
		} else {
			entry.downFile = filepath.Join(migrationsDir, name)
		}
		filesByVersion[version] = entry
	}

	versions := make([]string, 0, len(filesByVersion))
	for v := range filesByVersion {
		versions = append(versions, v)
	}
	sort.Strings(versions)

	var migrations []migration
	for _, v := range versions {
		entry := filesByVersion[v]

		var upSQL, downSQL string
		if entry.upFile != "" {
			data, err := os.ReadFile(entry.upFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read %s: %w", entry.upFile, err)
			}
			upSQL = string(data)
		}
		if entry.downFile != "" {
			data, err := os.ReadFile(entry.downFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read %s: %w", entry.downFile, err)
			}
			downSQL = string(data)
		}

		migrations = append(migrations, migration{
			version: v,
			name:    entry.name,
			upSQL:   upSQL,
			downSQL: downSQL,
		})
	}

	return migrations, nil
}

func getAppliedVersions(ctx context.Context, conn driver.Conn) (map[string]bool, error) {
	rows, err := conn.Query(ctx, "SELECT version FROM link4it.schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		applied[version] = true
	}
	return applied, nil
}

func runUp(ctx context.Context, conn driver.Conn, migrations []migration) {
	applied, err := getAppliedVersions(ctx, conn)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	for _, m := range migrations {
		if applied[m.version] {
			fmt.Printf("⊘ %s_%s already applied, skipping\n", m.version, m.name)
			continue
		}

		if m.upSQL == "" {
			fmt.Printf("⚠  %s_%s has no up migration, skipping\n", m.version, m.name)
			continue
		}

		fmt.Printf("Running migration: %s_%s\n", m.version, m.name)

		stmtCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		if err := execStatements(stmtCtx, conn, m.upSQL); err != nil {
			cancel()
			fmt.Printf("✗ Migration %s_%s failed: %v\n", m.version, m.name, err)
			os.Exit(1)
		}
		cancel()

		if err := conn.Exec(ctx, "INSERT INTO link4it.schema_migrations (version) VALUES (?)", m.version); err != nil {
			fmt.Printf("✗ Failed to record migration %s: %v\n", m.version, err)
			os.Exit(1)
		}

		fmt.Printf("✓ Migration %s_%s applied successfully\n", m.version, m.name)
	}
}

func runDown(ctx context.Context, conn driver.Conn, migrations []migration) {
	applied, err := getAppliedVersions(ctx, conn)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	for i := len(migrations) - 1; i >= 0; i-- {
		m := migrations[i]
		if !applied[m.version] {
			continue
		}

		if m.downSQL == "" {
			fmt.Printf("⚠  %s_%s has no down migration, skipping\n", m.version, m.name)
			continue
		}

		fmt.Printf("Rolling back: %s_%s\n", m.version, m.name)

		stmtCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		if err := execStatements(stmtCtx, conn, m.downSQL); err != nil {
			cancel()
			fmt.Printf("✗ Rollback %s_%s failed: %v\n", m.version, m.name, err)
			os.Exit(1)
		}
		cancel()

		if err := conn.Exec(ctx, "ALTER TABLE link4it.schema_migrations DELETE WHERE version = ?", m.version); err != nil {
			fmt.Printf("✗ Failed to remove migration record %s: %v\n", m.version, err)
			os.Exit(1)
		}

		fmt.Printf("✓ Rollback %s_%s applied successfully\n", m.version, m.name)
		return
	}

	fmt.Println("No applied migrations to roll back")
}

func runStatus(ctx context.Context, conn driver.Conn, migrations []migration) {
	applied, err := getAppliedVersions(ctx, conn)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Migration Status:")
	fmt.Println("=================")

	for _, m := range migrations {
		status := "Pending"
		if applied[m.version] {
			status = "Applied"
		}
		fmt.Printf("  %s_%-40s %s\n", m.version, m.name, status)
	}
}

func execStatements(ctx context.Context, conn driver.Conn, sql string) error {
	statements := splitStatements(sql)
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if err := conn.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("failed to execute: %s: %w", truncateSQL(stmt, 80), err)
		}
	}
	return nil
}

func splitStatements(sql string) []string {
	var statements []string
	current := strings.Builder{}
	for _, line := range strings.Split(sql, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		current.WriteString(line)
		current.WriteString("\n")
		if strings.HasSuffix(strings.TrimSpace(line), ";") {
			statements = append(statements, current.String())
			current.Reset()
		}
	}
	if current.Len() > 0 {
		remaining := strings.TrimSpace(current.String())
		if remaining != "" {
			statements = append(statements, remaining)
		}
	}
	return statements
}

func truncateSQL(sql string, maxLen int) string {
	oneLine := strings.ReplaceAll(strings.ReplaceAll(sql, "\n", " "), "\r", "")
	oneLine = strings.TrimSpace(oneLine)
	if len(oneLine) <= maxLen {
		return oneLine
	}
	return oneLine[:maxLen] + "..."
}
