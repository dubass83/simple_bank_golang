package db

import (
	"context"
	"database/sql"
	// "github.com/dubass83/simplebank/worker"
)

// VerifyEmailTxParams struct with arguments for VerifyEmailTx function
type VerifyEmailTxParams struct {
	ID         int64
	SecretCode string
}

// VerifyEmailTxResults struct with results from VerifyEmailTx function
type VerifyEmailTxResult struct {
	User        User
	VerifyEmail VerifyEmail
}

// VerifyEmailTx public method to crete new VerifyEmail in transaction
func (store *SQLStore) VerifyEmailTx(
	ctx context.Context,
	arg VerifyEmailTxParams) (VerifyEmailTxResult, error) {

	var result VerifyEmailTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.VerifyEmail, err = q.UpdateVerifyEmail(ctx, UpdateVerifyEmailParams{
			ID:         arg.ID,
			SecretCode: arg.SecretCode,
		})
		if err != nil {
			return err
		}
		result.User, err = q.UpdateUser(ctx, UpdateUserParams{
			Username: result.VerifyEmail.Username,
			IsEmailVerified: sql.NullBool{
				Bool:  true,
				Valid: true,
			},
		})

		return err
	})

	return result, err
}
