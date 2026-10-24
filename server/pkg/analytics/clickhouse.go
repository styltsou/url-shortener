package analytics

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
	"github.com/styltsou/url-shortener/server/pkg/logger"
	"go.uber.org/zap"
)

type ClickEvent struct {
	EventID   uuid.UUID `json:"event_id"`
	LinkID    uuid.UUID `json:"link_id"`
	Timestamp time.Time `json:"timestamp"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Referrer  string    `json:"referrer"`
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
}

type ClickRecorder interface {
	RecordClick(ctx context.Context, event ClickEvent)
}

type ClientInterface interface {
	ClickRecorder
	GetLinkAnalytics(ctx context.Context, linkID uuid.UUID, since, until time.Time) (*LinkAnalytics, error)
	GetUserAnalytics(ctx context.Context, linkIDs []uuid.UUID, since, until time.Time) (*UserAnalytics, error)
	DeliveryReporter
	Close() error
}

var ErrUnavailable = errors.New("analytics unavailable")

// Old events have no event ID; queued events use a stable ID across redeliveries.
const uniqueClickCount = "countIf(isNull(event_id)) + uniqExactIf(event_id, isNotNull(event_id))"

const (
	DeliveryOK       = "ok"
	DeliveryDisabled = "disabled"
	DeliveryDegraded = "degraded"
)

// DeliveryStats describes click delivery for this process since it started.
// Submitted includes clicks that could not be queued or written.
// Written means accepted by RabbitMQ when configured, or inserted into ClickHouse otherwise.
type DeliveryStats struct {
	Status          string
	Submitted       uint64
	Written         uint64
	LostQueueFull   uint64
	LostInsert      uint64
	LostPublish     uint64
	LostShutdown    uint64
	LostClosed      uint64
	LostUnavailable uint64
	Queued          int
}

func (s DeliveryStats) Lost() uint64 {
	return s.LostQueueFull + s.LostInsert + s.LostPublish + s.LostShutdown + s.LostClosed + s.LostUnavailable
}

type DeliveryReporter interface {
	DeliveryStats() DeliveryStats
}

type nopClient struct {
	logger    logger.Logger
	status    string
	submitted atomic.Uint64
}

func (c *nopClient) RecordClick(_ context.Context, event ClickEvent) {
	lost := c.submitted.Add(1)
	if lost == 1 || lost&(lost-1) == 0 {
		fields := []zap.Field{zap.String("status", c.status), zap.Uint64("lost_unavailable_total", lost)}
		if c.status == DeliveryDisabled {
			c.logger.Info("Click analytics disabled; events are not recorded", fields...)
		} else {
			c.logger.Warn("Click analytics unavailable; events are not recorded", fields...)
		}
	}
}

func (c *nopClient) DeliveryStats() DeliveryStats {
	count := c.submitted.Load()
	return DeliveryStats{Status: c.status, Submitted: count, LostUnavailable: count}
}

func (c *nopClient) GetLinkAnalytics(_ context.Context, _ uuid.UUID, _, _ time.Time) (*LinkAnalytics, error) {
	return nil, ErrUnavailable
}

func (c *nopClient) GetUserAnalytics(_ context.Context, _ []uuid.UUID, _, _ time.Time) (*UserAnalytics, error) {
	return nil, ErrUnavailable
}

func (c *nopClient) Close() error {
	if count := c.submitted.Load(); count > 0 {
		c.logger.Warn("Click analytics delivery summary",
			zap.String("status", c.status),
			zap.Uint64("submitted", count),
			zap.Uint64("lost_unavailable", count),
		)
	}
	return nil
}

type Client struct {
	conn             driver.Conn
	publisher        eventPublisher
	queryUnavailable bool
	logger           logger.Logger
	clickChan        chan clickRequest
	workersCtx       context.Context
	workersCancel    context.CancelFunc
	tableName        string
	mu               sync.Mutex
	closed           bool
	closeDone        chan struct{}
	closeErr         error
	workers          sync.WaitGroup
	submitted        atomic.Uint64
	written          atomic.Uint64
	lostQueueFull    atomic.Uint64
	lostInsert       atomic.Uint64
	lostPublish      atomic.Uint64
	lostShutdown     atomic.Uint64
	lostClosed       atomic.Uint64
}

func New(ctx context.Context, cfg Config, log logger.Logger) (ClientInterface, error) {
	if cfg.URL == "" && cfg.RabbitURL == "" {
		log.Info("ClickHouse URL not configured, analytics disabled")
		return &nopClient{logger: log, status: DeliveryDisabled}, nil
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

	var conn driver.Conn
	var connectionErr error
	if cfg.URL != "" {
		conn, connectionErr = clickhouse.Open(&clickhouse.Options{
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
		if connectionErr == nil {
			connectionErr = conn.Ping(ctx)
			if connectionErr != nil {
				_ = conn.Close()
				conn = nil
			}
		}
		if connectionErr != nil && cfg.RabbitURL == "" {
			return &nopClient{logger: log, status: DeliveryDegraded}, fmt.Errorf("ClickHouse unavailable: %w", connectionErr)
		}
	}

	if conn != nil {
		log.Info("ClickHouse connected successfully", zap.String("url", cfg.URL))
	}
	var publisher eventPublisher
	if cfg.RabbitURL != "" {
		publisher = newRabbitPublisher(cfg.RabbitURL)
	}

	workersCtx, workersCancel := context.WithCancel(context.Background())
	c := &Client{
		conn:             conn,
		publisher:        publisher,
		queryUnavailable: conn == nil,
		logger:           log,
		clickChan:        make(chan clickRequest, clickChanBuffer),
		workersCtx:       workersCtx,
		workersCancel:    workersCancel,
		tableName:        tableName,
		closeDone:        make(chan struct{}),
	}
	c.startWorkers()
	return c, connectionErr
}

func (c *Client) startWorkers() {
	for i := range numClickWorkers {
		c.workers.Add(1)
		go c.worker(i)
	}
}

func (c *Client) worker(id int) {
	defer c.workers.Done()
	for req := range c.clickChan {
		if c.workersCtx.Err() != nil {
			c.lostShutdown.Add(1)
			continue // The shutdown deadline elapsed; discard remaining queued events.
		}
		c.insertClick(c.workersCtx, req.event)
	}
}

func (c *Client) insertClick(ctx context.Context, event ClickEvent) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var err error
	if c.publisher != nil {
		err = c.publisher.Publish(ctx, event)
	} else {
		err = c.conn.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (link_id, timestamp, ip, user_agent, referrer)
		VALUES (?, ?, ?, ?, ?)
	`, c.tableName), event.LinkID, event.Timestamp, event.IP, event.UserAgent, event.Referrer)
	}
	if err != nil {
		if c.workersCtx.Err() != nil {
			c.lostShutdown.Add(1)
			return
		}
		var failed uint64
		failureKind := "insert"
		if c.publisher != nil {
			failed = c.lostPublish.Add(1)
			failureKind = "publish"
		} else {
			failed = c.lostInsert.Add(1)
		}
		c.logger.Warn("Failed to deliver click event",
			zap.String("link_id", event.LinkID.String()),
			zap.String("failure_kind", failureKind),
			zap.Uint64("lost_delivery_total", failed),
			zap.Uint64("submitted_total", c.submitted.Load()),
			zap.Error(err),
		)
		return
	}
	c.written.Add(1)
}

// InsertBatch writes a queue delivery as one ClickHouse insert. The caller
// acknowledges the queue messages only after this method succeeds.
func (c *Client) InsertBatch(ctx context.Context, events []ClickEvent) error {
	if len(events) == 0 {
		return nil
	}
	if c.conn == nil {
		return ErrUnavailable
	}
	batch, err := c.conn.PrepareBatch(ctx, fmt.Sprintf(`
		INSERT INTO %s (link_id, timestamp, ip, user_agent, referrer, event_id)
	`, c.tableName))
	if err != nil {
		return fmt.Errorf("prepare click batch: %w", err)
	}
	for _, event := range events {
		if event.EventID == uuid.Nil {
			return fmt.Errorf("click event has no event ID")
		}
		if err := batch.Append(event.LinkID, event.Timestamp, event.IP, event.UserAgent, event.Referrer, event.EventID); err != nil {
			return fmt.Errorf("append click event: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("send click batch: %w", err)
	}
	return nil
}

func (c *Client) RecordClick(_ context.Context, event ClickEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.submitted.Add(1)
	if c.closed {
		c.lostClosed.Add(1)
		return
	}
	select {
	case c.clickChan <- clickRequest{event: event}:
	default:
		lost := c.lostQueueFull.Add(1)
		c.logger.Warn("Click event dropped — worker queue full",
			zap.String("link_id", event.LinkID.String()),
			zap.Uint64("lost_queue_full_total", lost),
			zap.Uint64("submitted_total", c.submitted.Load()),
		)
	}
}

func (c *Client) DeliveryStats() DeliveryStats {
	stats := DeliveryStats{
		Submitted:     c.submitted.Load(),
		Written:       c.written.Load(),
		LostQueueFull: c.lostQueueFull.Load(),
		LostInsert:    c.lostInsert.Load(),
		LostPublish:   c.lostPublish.Load(),
		LostShutdown:  c.lostShutdown.Load(),
		LostClosed:    c.lostClosed.Load(),
		Queued:        len(c.clickChan),
	}
	stats.Status = DeliveryOK
	if stats.Lost() > 0 || c.queryUnavailable {
		stats.Status = DeliveryDegraded
	}
	return stats
}

func (c *Client) GetLinkAnalytics(ctx context.Context, linkID uuid.UUID, since, until time.Time) (*LinkAnalytics, error) {
	if c.conn == nil {
		return nil, ErrUnavailable
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
		ClicksOverTime: fillDailyClicks(clicksOverTime, since, until),
		TopReferrers:   topReferrers,
		TopUserAgents:  topUserAgents,
	}, nil
}

func (c *Client) getTotalClicks(ctx context.Context, linkID uuid.UUID, since, until time.Time) (int, error) {
	rows, err := c.conn.Query(ctx, fmt.Sprintf(`
		SELECT %s FROM %s
		WHERE link_id = ? AND timestamp >= ? AND timestamp <= ?
	`, uniqueClickCount, c.tableName), linkID, since, until)
	if err != nil {
		return 0, fmt.Errorf("failed to query total clicks: %w", err)
	}
	defer rows.Close()

	var total uint64
	for rows.Next() {
		if err := rows.Scan(&total); err != nil {
			return 0, fmt.Errorf("failed to scan total clicks: %w", err)
		}
	}
	return clickCountToInt(total)
}

func (c *Client) getClicksOverTime(ctx context.Context, linkID uuid.UUID, since, until time.Time) ([]ClicksOverTime, error) {
	rows, err := c.conn.Query(ctx, fmt.Sprintf(`
		SELECT toDate(timestamp, 'UTC') AS date, %s AS clicks
		FROM %s
		WHERE link_id = ? AND timestamp >= ? AND timestamp <= ?
		GROUP BY date
		ORDER BY date
	`, uniqueClickCount, c.tableName), linkID, since, until)
	if err != nil {
		return nil, fmt.Errorf("failed to query clicks over time: %w", err)
	}
	defer rows.Close()

	var result []ClicksOverTime
	for rows.Next() {
		var stat ClicksOverTime
		var clicks uint64
		if err := rows.Scan(&stat.Date, &clicks); err != nil {
			return nil, fmt.Errorf("failed to scan clicks over time: %w", err)
		}
		stat.Clicks, err = clickCountToInt(clicks)
		if err != nil {
			return nil, err
		}
		result = append(result, stat)
	}
	return result, nil
}

func (c *Client) getTopReferrers(ctx context.Context, linkID uuid.UUID, since, until time.Time) ([]ReferrerStat, error) {
	rows, err := c.conn.Query(ctx, fmt.Sprintf(`
		SELECT referrer, %s AS clicks
		FROM %s
		WHERE link_id = ? AND timestamp >= ? AND timestamp <= ?
		GROUP BY referrer
		ORDER BY clicks DESC
		LIMIT 10
	`, uniqueClickCount, c.tableName), linkID, since, until)
	if err != nil {
		return nil, fmt.Errorf("failed to query top referrers: %w", err)
	}
	defer rows.Close()

	var result []ReferrerStat
	for rows.Next() {
		var stat ReferrerStat
		var clicks uint64
		if err := rows.Scan(&stat.Referrer, &clicks); err != nil {
			return nil, fmt.Errorf("failed to scan referrer stat: %w", err)
		}
		stat.Clicks, err = clickCountToInt(clicks)
		if err != nil {
			return nil, err
		}
		result = append(result, stat)
	}
	return result, nil
}

func (c *Client) getTopUserAgents(ctx context.Context, linkID uuid.UUID, since, until time.Time) ([]UserAgentStat, error) {
	rows, err := c.conn.Query(ctx, fmt.Sprintf(`
		SELECT user_agent, %s AS clicks
		FROM %s
		WHERE link_id = ? AND timestamp >= ? AND timestamp <= ?
		GROUP BY user_agent
		ORDER BY clicks DESC
		LIMIT 10
	`, uniqueClickCount, c.tableName), linkID, since, until)
	if err != nil {
		return nil, fmt.Errorf("failed to query top user agents: %w", err)
	}
	defer rows.Close()

	var result []UserAgentStat
	for rows.Next() {
		var stat UserAgentStat
		var clicks uint64
		if err := rows.Scan(&stat.UserAgent, &clicks); err != nil {
			return nil, fmt.Errorf("failed to scan user agent stat: %w", err)
		}
		stat.Clicks, err = clickCountToInt(clicks)
		if err != nil {
			return nil, err
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
	if c.conn == nil {
		return nil, ErrUnavailable
	}
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
		ClicksOverTime: fillDailyClicks(clicksOverTime, since, until),
	}, nil
}

func (c *Client) getUserTotalClicks(ctx context.Context, linkIDs []uuid.UUID, since, until time.Time) (int, error) {
	rows, err := c.conn.Query(ctx, fmt.Sprintf(`
		SELECT %s FROM %s
		WHERE link_id = ANY(?) AND timestamp >= ? AND timestamp <= ?
	`, uniqueClickCount, c.tableName), linkIDs, since, until)
	if err != nil {
		return 0, fmt.Errorf("failed to query user total clicks: %w", err)
	}
	defer rows.Close()

	var total uint64
	for rows.Next() {
		if err := rows.Scan(&total); err != nil {
			return 0, fmt.Errorf("failed to scan user total clicks: %w", err)
		}
	}
	return clickCountToInt(total)
}

func (c *Client) getUserClicksOverTime(ctx context.Context, linkIDs []uuid.UUID, since, until time.Time) ([]ClicksOverTime, error) {
	rows, err := c.conn.Query(ctx, fmt.Sprintf(`
		SELECT toDate(timestamp, 'UTC') AS date, %s AS clicks
		FROM %s
		WHERE link_id = ANY(?) AND timestamp >= ? AND timestamp <= ?
		GROUP BY date
		ORDER BY date
	`, uniqueClickCount, c.tableName), linkIDs, since, until)
	if err != nil {
		return nil, fmt.Errorf("failed to query user clicks over time: %w", err)
	}
	defer rows.Close()

	var result []ClicksOverTime
	for rows.Next() {
		var stat ClicksOverTime
		var clicks uint64
		if err := rows.Scan(&stat.Date, &clicks); err != nil {
			return nil, fmt.Errorf("failed to scan user clicks over time: %w", err)
		}
		stat.Clicks, err = clickCountToInt(clicks)
		if err != nil {
			return nil, err
		}
		result = append(result, stat)
	}
	return result, nil
}

func clickCountToInt(count uint64) (int, error) {
	if count > uint64(^uint(0)>>1) {
		return 0, fmt.Errorf("click count %d exceeds the supported integer range", count)
	}
	return int(count), nil
}

// fillDailyClicks gives the chart one point per UTC date in the reporting window.
func fillDailyClicks(rows []ClicksOverTime, since, until time.Time) []ClicksOverTime {
	byDate := make(map[string]int, len(rows))
	for _, row := range rows {
		byDate[row.Date.Format("2006-01-02")] += row.Clicks
	}
	start := since.UTC()
	day := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	end := until.UTC()
	result := make([]ClicksOverTime, 0, 31)
	for !day.After(end) {
		result = append(result, ClicksOverTime{Date: day, Clicks: byDate[day.Format("2006-01-02")]})
		day = day.AddDate(0, 0, 1)
	}
	return result
}

func (c *Client) Close() error {
	c.mu.Lock()
	if c.closed {
		done := c.closeDone
		c.mu.Unlock()
		<-done
		return c.closeErr
	}
	c.closed = true
	close(c.clickChan)
	c.mu.Unlock()

	// Let the workers consume the queue; cancel in-flight inserts at the deadline.
	timer := time.AfterFunc(5*time.Second, c.workersCancel)
	c.workers.Wait()
	timer.Stop()
	c.workersCancel()
	stats := c.DeliveryStats()
	c.logger.Info("Click analytics delivery summary",
		zap.Uint64("submitted", stats.Submitted),
		zap.Uint64("written", stats.Written),
		zap.Uint64("lost_queue_full", stats.LostQueueFull),
		zap.Uint64("lost_insert", stats.LostInsert),
		zap.Uint64("lost_publish", stats.LostPublish),
		zap.Uint64("lost_shutdown", stats.LostShutdown),
		zap.Uint64("lost_closed", stats.LostClosed),
	)
	if c.publisher != nil {
		c.closeErr = c.publisher.Close()
	}
	if c.conn != nil {
		if err := c.conn.Close(); c.closeErr == nil {
			c.closeErr = err
		}
	}
	close(c.closeDone)
	return c.closeErr
}
