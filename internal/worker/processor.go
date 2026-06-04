package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/vivek6201/biolynq/internal/utils"
	"gorm.io/gorm"
)

type TaskProcessor interface {
	Start() error
	ProcessTaskSendEmail(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server      *asynq.Server
	emailSender *utils.EmailSender
	db          *gorm.DB
}

func (r *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskSendEmail, r.ProcessTaskSendEmail)

	return r.server.Run(mux)
}

func NewRedisTaskProcessor(opts asynq.RedisClientOpt, emailSender *utils.EmailSender, db *gorm.DB) TaskProcessor {
	server := asynq.NewServer(
		opts,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"default": 10,
			},
		},
	)

	return &RedisTaskProcessor{
		server:      server,
		emailSender: emailSender,
		db:          db,
	}
}

func (r *RedisTaskProcessor) ProcessTaskSendEmail(ctx context.Context, task *asynq.Task) error {
	var payload SendEmailPayload

	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("Error unmarshalling email payload: %v", err)
	}

	params := &utils.EmailParams{
		To:      payload.To,
		From:    payload.From,
		Subject: payload.Subject,
		Content: payload.Content,
	}

	if err := r.emailSender.SendEmail(params); err != nil {
		return fmt.Errorf("Error sending email: %v", err)
	}
	fmt.Println("Email sent successfully")

	return nil
}
