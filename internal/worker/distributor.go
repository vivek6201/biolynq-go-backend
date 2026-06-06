package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

type TaskDistributor interface {
	DistributeTaskSendEmail(ctx context.Context, payload *SendEmailPayload, opts ...asynq.Option) error
	DistributeTaskRecordEvent(ctx context.Context, payload *RecordEventPayload, opts ...asynq.Option) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(redisOpts asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOpts)

	return &RedisTaskDistributor{client: client}
}

func (d *RedisTaskDistributor) DistributeTaskSendEmail(ctx context.Context, payload *SendEmailPayload, opts ...asynq.Option) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Error marshaling payload: %v", err)
	}

	task := asynq.NewTask(TaskSendEmail, jsonPayload, opts...)

	info, err := d.client.EnqueueContext(ctx, task)

	if err != nil {
		return fmt.Errorf("Error enqueuing task: %v", err)
	}

	fmt.Printf("Task enqueued: type %s, queue %s, id %s\n", info.Type, info.Queue, info.ID)
	return nil
}

func (d *RedisTaskDistributor) DistributeTaskRecordEvent(ctx context.Context, payload *RecordEventPayload, opts ...asynq.Option) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Error marshaling payload: %v", err)
	}

	task := asynq.NewTask(TaskRecordEvent, jsonPayload, opts...)

	info, err := d.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("Error enqueuing task: %v", err)
	}

	fmt.Printf("Task enqueued: type %s, queue %s, id %s\n", info.Type, info.Queue, info.ID)
	return nil
}
