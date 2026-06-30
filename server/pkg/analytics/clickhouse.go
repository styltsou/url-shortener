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

const (
	clickChanBuffer = 1000
	numClickWorkers = 2
)

type clickRequest struct {
	event ClickEvent
	ctx   context.Context
}

type ClientInterface interface {
	RecordClick(ctx context.Context, event ClickEvent)
	GetLinkAnalytics(ctx context.Context, linkID uuid.UUID, since, until time.Time) (*LinkAnalytics, error)
	GetUserAnalytics(ctx context.Context, linkIDs []uuid.UUID, since, until time.Time) (*UserAnalytics, error)
	Close() error
}

type nopClient struct {
	logger logger.Logger
}

func (c *nopClient) RecordClick(_ context.Context, event ClickEvent) {
	// no-op
}

func (c *nopClient) GetLinkAnalytics(_ context.Context, _ uuid.UUID, _, _ time.Time) (*LinkAnalytics, error) {
	return &LinkAnalytics{}, nil
}

func (c *nopClient) GetUserAnalytics(_ context.Context, _ []uuid.UUID, _, _ time.Time) (*UserAnalytics, error) {
	return &UserAnalytics{}, nil
}

func (c *nopClient) Close() error {
	return nil
}

type Client struct {
	conn          driver.Conn
	logger        logger.Logger
	clickChan     chan clickRequest
	workersCtx    context.Context
	workersCancel context.CancelFunc
	tableName     string
}

func New(ctx context.Context, cfg Config, log logger.Logger) (ClientInterface, error) {
	if cfg.URL == "" {
		log.Info("ClickHouse URL not configured, analytics disabled")
		return &nopClient{logger: log}, nil
	}

	dialTimeout := cfg.DialTimeout
	if dialTimeout == 0 {
		dialTimeout = 5 * time.Second
	}
	maxOpenConns := cfg.MaxOpenConns
	if maxOpenConns == 0 {
		maxOpenConns = 5
	}
	maxIdleConns := cfg.MaxIdleConns
	if maxIdleConns == 0 {
		maxIdleConns = 2
	}
	connMaxLifetime := cfg.ConnMaxLifetime
	if connMaxLifetime == 0 {
		connMaxLifetime = 5 * time.Minute
	}

	tableName := cfg.TableName
	if tableName == "" {
		tableName = "link4it.click_events"
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.URL},
		Auth: clickhouse.Auth{
			Username: cfg.Username,
			Password: cfg.Password,
		},
		DialTimeout:     dialTimeout,
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
	})
	if err != nil {
		return &nopClient{logger: log}, fmt.Errorf("failed to open ClickHouse connection: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		conn.Close()
		return &nopClient{logger: log}, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	if err := conn.Exec(ctx, fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			link_id    UUID,
			timestamp  DateTime DEFAULT now(),
			ip         String,
			user_agent String,
			referrer   String
		) ENGINE = MergeTree()
		ORDER BY (link_id, timestamp)
	`, tableName)); err != nil {
		conn.Close()
		return &nopClient{logger: log}, fmt.Errorf("failed to create click_events table: %w", err)
	}

	log.Info("ClickHouse connected successfully",
		zap.String("url", cfg.URL),
	)

	workersCtx, workersCancel := context.WithCancel(context.Background())
	c := &Client{
		conn:          conn,
		logger:        log,
		clickChan:     make(chan clickRequest, clickChanBuffer),
		workersCtx:    workersCtx,
		workersCancel: workersCancel,
		tableName:     tableName,
	}
	c.startWorkers()
	return c, nil
}

func (c *Client) startWorkers() {
	for i := range numClickWorkers {
		go c.worker(i)
	}
}

func (c *Client) worker(id int) {
	for {
		select {
		case <-c.workersCtx.Done():
			return
		case req := <-c.clickChan:
			c.insertClick(req.ctx, req.event)
		}
	}
}

func (c *Client) insertClick(ctx context.Context, event ClickEvent) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := c.conn.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (link_id, timestamp, ip, user_agent, referrer)
		VALUES (?, ?, ?, ?, ?)
	`, c.tableName), event.LinkID, event.Timestamp, event.IP, event.UserAgent, event.Referrer); err != nil {
		c.logger.Warn("Failed to record click event",
			zap.String("link_id", event.LinkID.String()),
			zap.Error(err),
		)
	}
}

func (c *Client) RecordClick(ctx context.Context, event ClickEvent) {
	select {
	case c.clickChan <- clickRequest{event: event, ctx: ctx}:
	default:
		c.logger.Warn("Click event dropped — worker queue full",
			zap.String("link_id", event.LinkID.String()),
		)
	}
}

func (c *Client) GetLinkAnalytics(ctx context.Context, linkID uuid.UUID, since, until time.Time) (*LinkAnalytics, error) {
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
	rows, err := c.conn.Query(ctx, fmt.Sprintf(`
		SELECT count() FROM %s
		WHERE link_id = ? AND timestamp >= ? AND timestamp <= ?
	`, c.tableName), linkID, since, until)
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
	rows, err := c.conn.Query(ctx, fmt.Sprintf(`
		SELECT toDate(timestamp) AS date, count() AS clicks
		FROM %s
		WHERE link_id = ? AND timestamp >= ? AND timestamp <= ?
		GROUP BY date
		ORDER BY date
	`, c.tableName), linkID, since, until)
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
	rows, err := c.conn.Query(ctx, fmt.Sprintf(`
		SELECT referrer, count() AS clicks
		FROM %s
		WHERE link_id = ? AND timestamp >= ? AND timestamp <= ?
		GROUP BY referrer
		ORDER BY clicks DESC
		LIMIT 10
	`, c.tableName), linkID, since, until)
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
	rows, err := c.conn.Query(ctx, fmt.Sprintf(`
		SELECT user_agent, count() AS clicks
		FROM %s
		WHERE link_id = ? AND timestamp >= ? AND timestamp <= ?
		GROUP BY user_agent
		ORDER BY clicks DESC
		LIMIT 10
	`, c.tableName), linkID, since, until)
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

type UserAnalytics struct {
	TotalClicks    int              `json:"total_clicks"`
	ClicksOverTime []ClicksOverTime `json:"clicks_over_time"`
}

func (c *Client) GetUserAnalytics(ctx context.Context, linkIDs []uuid.UUID, since, until time.Time) (*UserAnalytics, error) {
	if len(linkIDs) == 0 {
		return &UserAnalytics{}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	totalClicks, err := c.getUserTotalClicks(ctx, linkIDs, since, until)
	if err != nil {
		return nil, err
	}

	clicksOverTime, err := c.getUserClicksOverTime(ctx, linkIDs, since, until)
	if err != nil {
		return nil, err
	}

	return &UserAnalytics{
		TotalClicks:    totalClicks,
		ClicksOverTime: clicksOverTime,
	}, nil
}

func (c *Client) getUserTotalClicks(ctx context.Context, linkIDs []uuid.UUID, since, until time.Time) (int, error) {
	rows, err := c.conn.Query(ctx, fmt.Sprintf(`
		SELECT count() FROM %s
		WHERE link_id = ANY(?) AND timestamp >= ? AND timestamp <= ?
	`, c.tableName), linkIDs, since, until)
	if err != nil {
		return 0, fmt.Errorf("failed to query user total clicks: %w", err)
	}
	defer rows.Close()

	var total int
	for rows.Next() {
		if err := rows.Scan(&total); err != nil {
			return 0, fmt.Errorf("failed to scan user total clicks: %w", err)
		}
	}
	return total, nil
}

func (c *Client) getUserClicksOverTime(ctx context.Context, linkIDs []uuid.UUID, since, until time.Time) ([]ClicksOverTime, error) {
	rows, err := c.conn.Query(ctx, fmt.Sprintf(`
		SELECT toDate(timestamp) AS date, count() AS clicks
		FROM %s
		WHERE link_id = ANY(?) AND timestamp >= ? AND timestamp <= ?
		GROUP BY date
		ORDER BY date
	`, c.tableName), linkIDs, since, until)
	if err != nil {
		return nil, fmt.Errorf("failed to query user clicks over time: %w", err)
	}
	defer rows.Close()

	var result []ClicksOverTime
	for rows.Next() {
		var stat ClicksOverTime
		if err := rows.Scan(&stat.Date, &stat.Clicks); err != nil {
			return nil, fmt.Errorf("failed to scan user clicks over time: %w", err)
		}
		result = append(result, stat)
	}
	return result, nil
}

func (c *Client) Close() error {
	c.workersCancel()

	drainCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if len(c.clickChan) == 0 {
				return c.conn.Close()
			}
		case <-drainCtx.Done():
			remaining := len(c.clickChan)
			if remaining > 0 {
				c.logger.Warn("Click worker pool drained with remaining events",
					zap.Int("dropped_events", remaining),
				)
			}
			return c.conn.Close()
		}
	}
}
