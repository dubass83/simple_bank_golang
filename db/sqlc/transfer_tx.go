package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

// TransferTxParams struct with arguments for TransferTx function
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Ammount       int64 `json:"ammount"`
}

// TransferTxResults struct with results from TransferTx function
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

// TransferTx public method to crete new transfer in transaction
func (store *SQLStore) TransferTx(
	ctx context.Context,
	arg TransferTxParams) (TransferTxResult, error) {

	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: pgtype.Int8{
				Int64: arg.FromAccountID,
				Valid: true,
			},
			ToAccountID: pgtype.Int8{
				Int64: arg.ToAccountID,
				Valid: true,
			},
			Amount: arg.Ammount,
		})
		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: pgtype.Int8{
				Int64: arg.FromAccountID,
				Valid: true,
			},
			Amount: -arg.Ammount,
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: pgtype.Int8{
				Int64: arg.ToAccountID,
				Valid: true,
			},
			Amount: arg.Ammount,
		})
		if err != nil {
			return err
		}

		result.FromAccount, result.ToAccount, err = updateAcoountBalanceInOrder(
			ctx,
			q,
			arg.FromAccountID,
			arg.ToAccountID,
			arg.Ammount,
		)
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}

func updateAcoountBalanceInOrder(
	ctx context.Context,
	q *Queries,
	fromAccId, toAccId, amm int64) (fromAcc Account, toAcc Account, err error) {
	if toAccId > fromAccId {
		toAcc, err = q.AddToAccountBalance(ctx, AddToAccountBalanceParams{
			ID:     toAccId,
			Amount: amm,
		})
		if err != nil {
			return
		}
		fromAcc, err = q.AddToAccountBalance(ctx, AddToAccountBalanceParams{
			ID:     fromAccId,
			Amount: -amm,
		})
		if err != nil {
			return
		}
	} else {
		fromAcc, err = q.AddToAccountBalance(ctx, AddToAccountBalanceParams{
			ID:     fromAccId,
			Amount: -amm,
		})
		if err != nil {
			return
		}
		toAcc, err = q.AddToAccountBalance(ctx, AddToAccountBalanceParams{
			ID:     toAccId,
			Amount: amm,
		})
		if err != nil {
			return
		}
	}
	return
}
