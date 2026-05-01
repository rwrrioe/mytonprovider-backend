package redisstream

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"mytonprovider-backend/pkg/jobs"
)

type Publisher struct {
	rdb          *redis.Client
	streamPrefix string
	maxLen       int64
}

func NewPublisher(
	rdb *redis.Client,
	streamPrefix string,
	maxLen int64,
) *Publisher {
	return &Publisher{
		rdb:          rdb,
		streamPrefix: streamPrefix,
		maxLen:       maxLen,
	}
}

func (p *Publisher) StreamFor(cycleType string) string {
	return fmt.Sprintf("%s:cycle:%s", p.streamPrefix, cycleType)
}

func (p *Publisher) Trigger(
	ctx context.Context,
	cycleType string,
	hint json.RawMessage,
) (jobID string, err error) {
	const op = "redisstream.Publisher.Trigger"

	jobID = uuid.NewString()
	env := jobs.TriggerEnvelope{
		JobID:      jobID,
		Type:       cycleType,
		Hint:       hint,
		EnqueuedAt: time.Now(),
	}

	data, err := json.Marshal(env)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	args := &redis.XAddArgs{
		Stream: p.StreamFor(cycleType),
		Values: map[string]any{"data": data},
	}
	if p.maxLen > 0 {
		args.MaxLen = p.maxLen
		args.Approx = true
	}

	if err = p.rdb.XAdd(ctx, args).Err(); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return jobID, nil
}
