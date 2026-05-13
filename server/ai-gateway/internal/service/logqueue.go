package service

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

type LogEntry struct {
	UserId           int64
	ApiKeyId         int64
	ModelName        string
	PromptTokens     int
	CompletionTokens int
	Quota            float64
	RequestId        string
	ChannelId        int64
	ChannelName      string
}

type LogQueue struct {
	entries chan LogEntry
	done    chan struct{}
}

func NewLogQueue(bufferSize int) *LogQueue {
	if bufferSize <= 0 {
		bufferSize = 10000
	}
	return &LogQueue{
		entries: make(chan LogEntry, bufferSize),
		done:    make(chan struct{}),
	}
}

func (q *LogQueue) Start(ctx context.Context, batchSize int, interval time.Duration) {
	if batchSize <= 0 {
		batchSize = 50
	}
	if interval <= 0 {
		interval = 100 * time.Millisecond
	}

	go func() {
		batch := make([]LogEntry, 0, batchSize)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		flush := func() {
			if len(batch) == 0 {
				return
			}
			if err := q.batchInsert(ctx, batch); err != nil {
				glog.Errorf(ctx, "log queue batch insert failed: %v", err)
			}
			batch = batch[:0]
		}

		for {
			select {
			case entry := <-q.entries:
				batch = append(batch, entry)
				if len(batch) >= batchSize {
					flush()
				}
			case <-ticker.C:
				flush()
			case <-ctx.Done():
				flush()
				close(q.done)
				return
			}
		}
	}()
}

func (q *LogQueue) Stop() {
	close(q.entries)
	<-q.done
}

func (q *LogQueue) Push(entry LogEntry) {
	select {
	case q.entries <- entry:
	default:
		glog.Printf(nil, "log queue full, dropping entry: %s", entry.RequestId)
	}
}

func (q *LogQueue) batchInsert(ctx context.Context, batch []LogEntry) error {
	return g.DB().Model("usage_records").Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		for _, e := range batch {
			_, err := tx.Insert("usage_records", g.Map{
				"user_id":           e.UserId,
				"api_key_id":        e.ApiKeyId,
				"model_name":        e.ModelName,
				"prompt_tokens":     e.PromptTokens,
				"completion_tokens": e.CompletionTokens,
				"quota":             e.Quota,
				"request_id":        e.RequestId,
				"channel_id":        e.ChannelId,
				"channel_name":      e.ChannelName,
				"created_at":        time.Now(),
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

var LogQ = NewLogQueue(10000)
