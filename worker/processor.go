package worker

import (
	"context"

	"github.com/hibiken/asynq"
	db "github.com/mehtabghani/simplebank/db/sqlc"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

// This processor will pick up the task from Redis and process them.
type TaskProcessor interface {
	Start() error
	ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store  db.Store
	// mailer mail.EmailSender
}

// func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store, mailer mail.EmailSender) TaskProcessor {
func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store) TaskProcessor {

	// logger := NewLogger()
	// redis.SetLogger(logger)

	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			// Queues: map[string]int{
			// 	QueueCritical: 10,
			// 	QueueDefault:  5,
			// },
			// ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			// 	log.Error().Err(err).Str("type", task.Type()).
			// 		Bytes("payload", task.Payload()).Msg("process task failed")
			// }),
			// Logger: logger,
		},
	)

	return &RedisTaskProcessor{
		server: server,
		store:  store,
		// mailer: mailer,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskSendVerifyEmail, processor.ProcessTaskSendVerifyEmail)

	// IMP: if you have more task register here like above

	return processor.server.Start(mux)
}
