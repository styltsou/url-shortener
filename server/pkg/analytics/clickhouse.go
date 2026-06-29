package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
	"github.com/styltsou/url-shortener/server/pkg/logger"
	"go.uber.org/zap"
)

type ClickEvent struct {
	LinkID    uuid.UUID
	Timestamp time.Time
	IP        string
	UserAgent string
	Referrer  string
}

type ClicksOverTime struct {
	Date   time.Time `json:"date"`
	Clicks int       `json:"clicks"`
}

type ReferrerStat struct {
	Referrer string `json:"referrer"`
	Clicks   int    `json:"clicks"`
}

type UserAgentStat struct {
	UserAgent string `json:"user_agent"`
	Clicks    int    `json:"clicks"`
}

type LinkAnalytics struct {
	TotalClicks    int              `json:"total_clicks"`
	ClicksOverTime []ClicksOverTime `json:"clicks_over_time"`
	TopReferrers   []ReferrerStat   `json:"top_referrers"`
	TopUserAgents  []UserAgentStat  `json:"top_user_agents"`
}

type Client struct {
	conn   driver.Conn
	logger logger.Logger
}

func New(ctx context.Context, cfg Config, log logger.Logger) (*Client, error) {
	if cfg.URL == "" {
		log.Info("ClickHouse URL not configured, analytics disabled")
		return &Client{conn: nil, logger: log}, nil
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.URL},
		Auth: clickhouse.Auth{
			Username: cfg.Username,
			Password: cfg.Password,
		},
		DialTimeout:      5 * time.Second,
		MaxOpenConns:     5,
		MaxIdleConns:     2,
		ConnMaxLifetime:  time.Minute * 5,
	})
	if err != nil {
		return &Client{conn: nil, logger: log},
			fmt.Errorf("failed to open ClickHouse connection: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		conn.Close()
		return &Client{conn: nil, logger: log},
			fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	if err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS link4it.click_events (
			link_id    UUID,
			timestamp  DateTime DEFAULT now(),
			ip         String,
			user_agent String,
			referrer   String
		) ENGINE = MergeTree()
		ORDER BY (link_id, timestamp)
	`); err != nil {
		conn.Close()
		return &Client{conn: nil, logger: log},
			fmt.Errorf("failed to create click_events table: %w", err)
	}

	log.Info("ClickHouse connected successfully",
		zap.String("url", cfg.URL),
	)

	return &Client{conn: conn, logger: log}, nil
}

func (c *Client) RecordClick(ctx context.Context, event ClickEvent) {
	if c.conn == nil {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := c.conn.Exec(ctx, `
		INSERT INTO link4it.click_events (link_id, timestamp, ip, user_agent, referrer)
		VALUES (?, ?, ?, ?, ?)
	`, event.LinkID, event.Timestamp, event.IP, event.UserAgent, event.Referrer); err != nil {
		c.logger.Warn("Failed to record click event",
			zap.String("link_id", event.LinkID.String()),
			zap.Error(err),
		)
	}
}

func (c *Client) GetLinkAnalytics(ctx context.Context, linkID uuid.UUID, since, until time.Time) (*LinkAnalytics, error) {
	if c.conn == nil {
		return &LinkAnalytics{}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	totalClicks, err := c.getTotalClicks(ctx, linkID, since, until)
	if err != nil {
		return nil, err
	}

	clicksOverTime, err := c.getClicksOverTime(ctx, linkID, since, until)
	if err != nil {
		return nil, err
	}

	topReferrers, err := c.getTopReferrers(ctx, linkID, since, until)
	if err != nil {
		return nil, err
	}

	topUserAgents, err := c.getTopUserAgents(ctx, linkID, since, until)
	if err != nil {
		return nil, err
	}

	return &LinkAnalytics{
		TotalClicks:    totalClicks,
		ClicksOverTime: clicksOverTime,
		TopReferrers:   topReferrers,
		TopUserAgents:  topUserAgents,
	}, nil
}

func (c *Client) getTotalClicks(ctx context.Context, linkID uuid.UUID, since, until time.Time) (int, error) {
	rows, err := c.conn.Query(ctx, `
		SELECT count() FROM link4it.click_events
		WHERE link_id = ? AND timestamp >= ? AND timestamp <= ?
	`, linkID, since, until)
	if err != nil {
		return 0, fmt.Errorf("failed to query total clicks: %w", err)
	}
	defer rows.Close()

	var total int
	for rows.Next() {
		if err := rows.Scan(&total); err != nil {
			return 0, fmt.Errorf("failed to scan total clicks: %w", err)
		}
	}
	return total, nil
}

func (c *Client) getClicksOverTime(ctx context.Context, linkID uuid.UUID, since, until time.Time) ([]ClicksOverTime, error) {
	rows, err := c.conn.Query(ctx, `
		SELECT toDate(timestamp) AS date, count() AS clicks
		FROM link4it.click_events
		WHERE link_id = ? AND timestamp >= ? AND timestamp <= ?
		GROUP BY date
		ORDER BY date
	`, linkID, since, until)
	if err != nil {
		return nil, fmt.Errorf("failed to query clicks over time: %w", err)
	}
	defer rows.Close()

	var result []ClicksOverTime
	for rows.Next() {
		var stat ClicksOverTime
		if err := rows.Scan(&stat.Date, &stat.Clicks); err != nil {
			return nil, fmt.Errorf("failed to scan clicks over time: %w", err)
		}
		result = append(result, stat)
	}
	return result, nil
}

func (c *Client) getTopReferrers(ctx context.Context, linkID uuid.UUID, since, until time.Time) ([]ReferrerStat, error) {
	rows, err := c.conn.Query(ctx, `
		SELECT referrer, count() AS clicks
		FROM link4it.click_events
		WHERE link_id = ? AND timestamp >= ? AND timestamp <= ?
		GROUP BY referrer
		ORDER BY clicks DESC
		LIMIT 10
	`, linkID, since, until)
	if err != nil {
		return nil, fmt.Errorf("failed to query top referrers: %w", err)
	}
	defer rows.Close()

	var result []ReferrerStat
	for rows.Next() {
		var stat ReferrerStat
		if err := rows.Scan(&stat.Referrer, &stat.Clicks); err != nil {
			return nil, fmt.Errorf("failed to scan referrer stat: %w", err)
		}
		result = append(result, stat)
	}
	return result, nil
}

func (c *Client) getTopUserAgents(ctx context.Context, linkID uuid.UUID, since, until time.Time) ([]UserAgentStat, error) {
	rows, err := c.conn.Query(ctx, `
		SELECT user_agent, count() AS clicks
		FROM link4it.click_events
		WHERE link_id = ? AND timestamp >= ? AND timestamp <= ?
		GROUP BY user_agent
		ORDER BY clicks DESC
		LIMIT 10
	`, linkID, since, until)
	if err != nil {
		return nil, fmt.Errorf("failed to query top user agents: %w", err)
	}
	defer rows.Close()

	var result []UserAgentStat
	for rows.Next() {
		var stat UserAgentStat
		if err := rows.Scan(&stat.UserAgent, &stat.Clicks); err != nil {
			return nil, fmt.Errorf("failed to scan user agent stat: %w", err)
		}
		result = append(result, stat)
	}
	return result, nil
}

func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}
