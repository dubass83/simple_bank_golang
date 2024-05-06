package worker

import (
	"context"
	"encoding/json"
	"fmt"

	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/dubass83/simplebank/util"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const (
	TaskSendVerifyEmail = "task:send_verify_email"
	EmailBody           = `<h1>Hi there, %s!</h1></br>
	<p>This is wellcome message from simple bank dev project</p></br>
    <p>To verify email go to this <a href="%s">link!</a></p></br>
	<p>You can find source code <a href="https://github.com/dubass83/simple_bank_golang">here</a></p>`
)

type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}

func (distributor *RedisTaskDistributor) DestributeTaskSendVerifyEmail(
	ctx context.Context,
	payload *PayloadSendVerifyEmail,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed unmarshal payload %w", err)
	}
	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).
		Msg("enqueued task")
	return nil
}

func (processor *RedisTaskProcessor) ProcesTaskVerifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendVerifyEmail
	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return fmt.Errorf("failed unmarshal payload: %w", asynq.SkipRetry)
	}

	user, err := processor.store.GetUser(ctx, payload.Username)
	if err != nil {
		// if err == sql.ErrNoRows {
		// 	return fmt.Errorf("user does not exist: %w", asynq.SkipRetry)
		// }
		return fmt.Errorf("failed to get user: %w", err)
	}

	ve, err := processor.store.CreateVerifyEmail(ctx, db.CreateVerifyEmailParams{
		Username:   user.Username,
		Email:      user.Email,
		SecretCode: util.RandomString(32),
	})
	if err != nil {
		return fmt.Errorf("failed to create verify email: %w", err)
	}
	subject := "Hello from Simple Bank!"
	verifyURL := fmt.Sprintf("http://localhost:8080/v1/verify_email?id=%d&secret_code=%s", ve.ID, ve.SecretCode)
	content := fmt.Sprintf(EmailBody, user.Username, verifyURL)
	to := []string{"makssych@gmail.com", "makssych@outlook.com"}
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", user.Email).Msg("processed task")
	return processor.sender.SendEmail(subject, content, to, nil, nil, nil)
}
