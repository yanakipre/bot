package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dranikpg/gtrs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/encodingtooling"
	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/metrics"
)

type InnerAck gtrs.InnerAck

// GroupConsumerConfig provides basic configuration for GroupConsumer.
type GroupConsumerConfig struct {
	// StreamName what stream to read from.
	StreamName string `yaml:"stream_name"`
	// milliseconds to block before timing out. 0 means infinite
	Block encodingtooling.Duration `yaml:"block"`
	// maximum number of entries per request. 0 means not limited
	// We read Count from Redis at once,
	// and write them into a channel of BufferSize,
	// blocking further reads until channel empties.
	Count int64 `yaml:"count"`
	// how many entries to prefetch at most
	BufferSize uint `yaml:"buffer_size"`
	// AckBufferSize is the size of the buffer for ack requests.
	// the driver blocks when ack buffer is full.
	// Acks behave like a backpressure if called within consumption loop,
	// if you are geographically far away.
	// Consumption might be deadlocked according to the docs.
	AckBufferSize uint `yaml:"ack_buffer_size"`
}

func DefaultConsumerConfig(streamName string) GroupConsumerConfig {
	return GroupConsumerConfig{
		StreamName: streamName,
		Block:      encodingtooling.Duration{},
		Count:      100,
		BufferSize: 50,
		// Examples:
		// 1. Take Singapore: 200msec RTT, with 16 RPS we need 3,2 sec to ack everything that came.
		//
		// 2. If we have around 160000 events, one burst in 30min. In Singapore, with 200msec RTT
		// we need 533 minutes to ack them all. We'll never ack them all, staying locked,
		// waiting for acknowledgment in the main loop.
		// With Singapore, we're limited to 30*60/0,2 = 9000 events, 5 RPS.
		AckBufferSize: 9000,
	}
}

const redisName = "redis"

type message interface {
	// ProducedAt to know when message was produced on the producer side.
	ProducedAt() time.Time
}

// GroupConsumer is a consumer implementation using Redis-streams.
// https://redis.io/docs/data-types/streams/
type GroupConsumer[T message] struct {
	shutdownCtx context.Context
	*gtrs.GroupConsumer[T]
	mConsDelay     prometheus.Observer
	mConsTotal     prometheus.Counter
	mConsProcessed prometheus.Counter
}

func (c *GroupConsumer[T]) Close() []InnerAck {
	return lo.Map(c.GroupConsumer.Close(), func(item gtrs.InnerAck, _ int) InnerAck {
		return InnerAck(item)
	})
}

func (c *GroupConsumer[T]) Chan() <-chan gtrs.Message[T] {
	ctx := logger.WithName(c.shutdownCtx, "consumer")
	wrappedC := make(chan gtrs.Message[T])
	go func() {
		for msg := range c.GroupConsumer.Chan() {
			c.mConsTotal.Inc()
			switch msg.Err.(type) {
			case nil: // This interface-nil comparison in safe
				c.mConsDelay.Observe(time.Since(msg.Data.ProducedAt()).Seconds())
				wrappedC <- msg
			case StreamReadError:
				logger.Error(ctx, "error reading from redis message bus", zap.Error(msg.Err))
				return // last message in channel
			case StreamAckError:
				logger.Error(ctx,
					"ack failed",
					zap.Error(msg.Err),
					zap.String("msg_stream", msg.Stream),
					zap.String("msg_id", msg.ID))
			case StreamParseError:
				marshal, err := json.Marshal(msg.Data)
				if err != nil {
					logger.Error(ctx,
						"failed to marshall msg that failed to be parsed, not supposed to happen")
					continue
				}
				if len(marshal) > 10000 {
					marshal = marshal[:10000]
				}
				logger.Error(ctx, "failed to parse msg from redis message bus",
					zap.ByteString("msg_data", marshal),
					zap.String("msg_stream", msg.Stream),
					zap.String("msg_id", msg.ID))
				// still ACK the message,
				// we logged it and don't want it to stay in the stream forever.
				c.Ack(msg)
				return
			case error:
				logger.Error(ctx, "unknown error reading from redis message bus",
					zap.Error(msg.Err),
					zap.String("msg_stream", msg.Stream),
					zap.String("msg_id", msg.ID))
				return
			}
		}
		defer close(wrappedC)
	}()
	return wrappedC
}

// Ack avoid blocking the consumption loop in the underlying driver
func (c *GroupConsumer[T]) Ack(m gtrs.Message[T]) {
	c.mConsProcessed.Inc()
	c.GroupConsumer.Ack(m)
}

func NewGroupConsumer[T message](
	ctx context.Context,
	rdb redis.UniversalClient,
	group, name, lastID string,
	cfg GroupConsumerConfig,
) *GroupConsumer[T] {
	labels := []string{
		redisName,
		cfg.StreamName,
		group,
		name,
	}

	c := &GroupConsumer[T]{
		shutdownCtx: ctx,
		GroupConsumer: gtrs.NewGroupConsumer[T](
			ctx,
			rdb,
			group,
			name,
			cfg.StreamName,
			lastID,
			gtrs.GroupConsumerConfig{
				StreamConsumerConfig: gtrs.StreamConsumerConfig{
					Block:      cfg.Block.Duration,
					Count:      cfg.Count,
					BufferSize: cfg.BufferSize,
				},
				AckBufferSize: cfg.AckBufferSize,
			},
		),
		mConsDelay: metrics.MBusMessageConsumptionDelayDuration.WithLabelValues(
			labels...,
		),
		mConsTotal: metrics.MBusMessagesConsumedTotal.WithLabelValues(
			labels...,
		),
		mConsProcessed: metrics.MBusMessagesConsumedProcessed.WithLabelValues(
			labels...,
		),
	}

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fetched, pendingAcks, unprocessedAckErrs := c.GroupConsumer.Stats()
				metrics.MBusPendingAckMessages.WithLabelValues(
					labels...,
				).Set(float64(pendingAcks))
				metrics.MBusFetchedEntries.WithLabelValues(
					labels...,
				).Set(float64(fetched))
				metrics.MBusUnprocessedAckErrors.WithLabelValues(
					labels...,
				).Set(float64(unprocessedAckErrs))
			}
		}
	}()

	return c
}
