package analytics

import "time"

type Config struct {
	URL             string
	Username        string
	Password        string
	DialTimeout     time.Duration
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	TableName       string
}
