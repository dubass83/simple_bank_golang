package db

import (
	"context"
	// "github.com/dubass83/simplebank/worker"
)

// CreateUserTxParams struct with arguments for CreateUserTx function
type CreateUserTxParams struct {
	CreateUserParams
	AfterFunc func(User) error
}

// CreateUserTxResults struct with results from CreateUserTx function
type CreateUserTxResult struct {
	User
}

// CreateUserTx public method to crete new CreateUser in transaction
func (store *SQLStore) CreateUserTx(
	ctx context.Context,
	arg CreateUserTxParams) (CreateUserTxResult, error) {

	var result CreateUserTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.User, err = q.CreateUser(ctx, arg.CreateUserParams)
		if err != nil {
			return err
		}
		return arg.AfterFunc(result.User)
	})

	return result, err
}
