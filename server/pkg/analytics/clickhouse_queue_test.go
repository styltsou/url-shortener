package analytics

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
)

type fakeClickConn struct {
	driver.Conn
	inserts atomic.Int32
	closed  atomic.Bool
	onExec  func(context.Context)
	execErr error
}

func (c *fakeClickConn) Exec(ctx context.Context, _ string, _ ...any) error {
	if c.onExec != nil {
		c.onExec(ctx)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if c.execErr != nil {
		return c.execErr
	}
	c.inserts.Add(1)
	return nil
}
func (c *fakeClickConn) Close() error { c.closed.Store(true); return nil }

func testClickClient(t *testing.T, conn *fakeClickConn, capacity int) *Client {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	c := &Client{conn: conn, logger: newTestLogger(t), clickChan: make(chan clickRequest, capacity), workersCtx: ctx, workersCancel: cancel, closeDone: make(chan struct{}), tableName: "click_events"}
	c.startWorkers()
	return c
}

func TestClickSurvivesCanceledRequestAndCloseDrains(t *testing.T) {
	conn := &fakeClickConn{}
	workerCtx, workerCancel := context.WithCancel(context.Background())
	c := &Client{
		conn: conn, logger: newTestLogger(t), clickChan: make(chan clickRequest, 32),
		workersCtx: workerCtx, workersCancel: workerCancel, closeDone: make(chan struct{}), tableName: "click_events",
	}
	request, cancel := context.WithCancel(context.Background())
	for range 20 {
		c.RecordClick(request, ClickEvent{LinkID: uuid.New()})
	}
	cancel()
	c.startWorkers() // No insert can begin until the request is already canceled.
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if got := conn.inserts.Load(); got != 20 {
		t.Fatalf("inserted %d of 20 queued clicks", got)
	}
	if stats := c.DeliveryStats(); stats.Submitted != 20 || stats.Written != 20 || stats.Lost() != 0 {
		t.Fatalf("delivery stats = %+v, want 20 written and no loss", stats)
	}
	if !conn.closed.Load() {
		t.Fatal("connection not closed")
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	c.RecordClick(context.Background(), ClickEvent{LinkID: uuid.New()})
	if got := conn.inserts.Load(); got != 20 {
		t.Fatalf("accepted click after close: %d", got)
	}
	if stats := c.DeliveryStats(); stats.LostClosed != 1 {
		t.Fatalf("lost after close = %d, want 1", stats.LostClosed)
	}
}

func TestFailedClickInsertIsCounted(t *testing.T) {
	c := testClickClient(t, &fakeClickConn{execErr: errors.New("clickhouse unavailable")}, 1)
	c.RecordClick(context.Background(), ClickEvent{LinkID: uuid.New()})
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	stats := c.DeliveryStats()
	if stats.Submitted != 1 || stats.Written != 0 || stats.LostInsert != 1 || stats.Status != DeliveryDegraded {
		t.Fatalf("delivery stats = %+v, want one failed insert", stats)
	}
}

func TestClickQueueFullAndConcurrentClose(t *testing.T) {
	started := make(chan struct{}, numClickWorkers)
	conn := &fakeClickConn{onExec: func(ctx context.Context) {
		started <- struct{}{}
		<-ctx.Done()
	}}
	c := testClickClient(t, conn, 1)
	for range numClickWorkers {
		c.RecordClick(context.Background(), ClickEvent{})
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("worker never started")
		}
	}
	c.RecordClick(context.Background(), ClickEvent{})
	c.RecordClick(context.Background(), ClickEvent{}) // dropped, never blocks
	if len(c.clickChan) != 1 {
		t.Fatalf("queue length = %d", len(c.clickChan))
	}
	if stats := c.DeliveryStats(); stats.LostQueueFull != 1 || stats.Status != DeliveryDegraded {
		t.Fatalf("delivery stats = %+v, want one full-queue loss", stats)
	}
	var producers sync.WaitGroup
	for range 10 {
		producers.Add(1)
		go func() {
			defer producers.Done()
			for range 10 {
				c.RecordClick(context.Background(), ClickEvent{})
			}
		}()
	}
	done := make(chan struct{})
	go func() { _ = c.Close(); close(done) }()
	producers.Wait()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("close did not stop workers at deadline")
	}
	if !conn.closed.Load() {
		t.Fatal("connection was not closed")
	}
	stats := c.DeliveryStats()
	if stats.Submitted != stats.Written+stats.Lost() {
		t.Fatalf("delivery accounting incomplete: %+v", stats)
	}
}
