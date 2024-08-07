package worker

import (
	"context"

	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/dubass83/simplebank/mail"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"
)

type TaskProcessor interface {
	ProcesTaskVerifyEmail(ctx context.Context, task *asynq.Task) error
	Start() error
	Stop()
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store  db.Store
	sender mail.EmailSender
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store, sender mail.EmailSender) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 6,
				QueueDefault:  3,
				QueueLow:      1,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Error().Err(err).Str("type", task.Type()).
					Bytes("payload", task.Payload()).Msg("process task failed")
			}),
			Logger: NewLoger(),
		},
	)

	return &RedisTaskProcessor{
		server: server,
		store:  store,
		sender: sender,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskSendVerifyEmail, processor.ProcesTaskVerifyEmail)

	return processor.server.Start(mux)
}

func (processor *RedisTaskProcessor) Stop() {
	processor.server.Shutdown()
}
