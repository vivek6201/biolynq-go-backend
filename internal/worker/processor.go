package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/vivek6201/biolynq/internal/utils"
)

type TaskProcessor interface {
	Start() error
	RegisterHandler(taskType string, handler asynq.HandlerFunc)
}

type RedisTaskProcessor struct {
	server      *asynq.Server
	emailSender *utils.EmailSender
	mux         *asynq.ServeMux
}

func (r *RedisTaskProcessor) Start() error {
	return r.server.Run(r.mux)
}

func (r *RedisTaskProcessor) RegisterHandler(taskType string, handler asynq.HandlerFunc) {
	r.mux.HandleFunc(taskType, handler)
}

func NewRedisTaskProcessor(opts asynq.RedisClientOpt, emailSender *utils.EmailSender) TaskProcessor {
	server := asynq.NewServer(
		opts,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"default": 10,
			},
		},
	)
	mux := asynq.NewServeMux()

	processor := &RedisTaskProcessor{
		server:      server,
		emailSender: emailSender,
		mux:         mux,
	}

	mux.HandleFunc(TaskSendEmail, processor.ProcessTaskSendEmail)

	return processor
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
