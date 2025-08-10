package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log/slog"
	"math"
	"math/rand"
	"time"
	"wbL0/internal/config"
	"wbL0/internal/models"
	"wbL0/internal/service/orderService"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

const (
	maxProcessAttempts = 5
	maxCommitAttempts  = 3

	baseBackoff = 500 * time.Millisecond
	maxBackoff  = 10 * time.Second

	readErrorBackoff = 1 * time.Second
)

func calcBackoff(attempt int) time.Duration {
	backoff := float64(baseBackoff) * math.Pow(2, float64(attempt-1))
	if time.Duration(backoff) > maxBackoff {
		backoff = float64(maxBackoff)
	}
	jitter := rng.Float64() * (backoff / 4)
	return time.Duration(backoff + jitter)
}

func ConsumeMessage(ctx context.Context, cfg *config.Config, svc orderService.OrderServiceInterface, log *slog.Logger) error {
	const op = "kafka.consumeMessage"

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Kafka.Brokers,
		Topic:          cfg.Kafka.Topic,
		GroupID:        cfg.Kafka.GroupID,
		Partition:      cfg.Kafka.Partition,
		MinBytes:       10e3,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
	})
	defer func() {
		if err := reader.Close(); err != nil {
			log.Warn("kafka reader close error", "err", err)
		}
	}()

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, ctx.Err()) {
				log.Info("consumer context canceled")
				return nil
			}
			log.Error("error reading kafka message", "op", op, "err", err, "error_str", err.Error())

			select {
			case <-time.After(readErrorBackoff):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		var full models.FullOrder
		if err := json.Unmarshal(msg.Value, &full); err != nil {
			log.Error("invalid message format, skipping", "op", op, "err", err, "topic", msg.Topic, "partition", msg.Partition, "offset", msg.Offset)
			continue
		}

		var procErr error
		for attempt := 1; attempt <= maxProcessAttempts; attempt++ {
			if errors.Is(ctx.Err(), context.Canceled) {
				log.Info("context canceled while processing", "order_uid", full.Order.OrderUID)
				return ctx.Err()
			}

			procErr = svc.ProcessAndCache(ctx, &full)
			if procErr == nil {
				break
			}

			log.Warn("failed to process order, will retry", "op", op, "order_uid", full.Order.OrderUID, "attempt", attempt, "err", procErr.Error())
			fmt.Println(procErr)
			if attempt < maxProcessAttempts {
				sleep := calcBackoff(attempt)
				select {
				case <-time.After(sleep):
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}

		if procErr != nil {
			log.Error("failed to process order after retries, skipping message", "op", op, "order_uid", full.Order.OrderUID, "err", procErr.Error())
			fmt.Println(procErr)
			continue
		}

		var commitErr error
		for attempt := 1; attempt <= maxCommitAttempts; attempt++ {
			commitErr = reader.CommitMessages(ctx, msg)
			if commitErr == nil {
				break
			}

			if errors.Is(ctx.Err(), context.Canceled) {
				log.Info("context canceled while committing", "order_uid", full.Order.OrderUID)
				return ctx.Err()
			}

			log.Warn("failed to commit message, will retry", "op", op, "order_uid", full.Order.OrderUID, "attempt", attempt, "err", commitErr.Error())
			fmt.Println(commitErr)
			if attempt < maxCommitAttempts {
				sleep := calcBackoff(attempt)
				select {
				case <-time.After(sleep):
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
		if commitErr != nil {
			log.Error("failed to commit message after retries", "op", op, "order_uid", full.Order.OrderUID, "err", commitErr.Error())
		}
	}
}
