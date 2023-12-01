package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		Queries: New(db),
		db:      db,
	}
}

// execTx private method which will run transaction
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rollback err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

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
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: sql.NullInt64{
				Int64: arg.FromAccountID,
				Valid: true,
			},
			ToAccountID: sql.NullInt64{
				Int64: arg.ToAccountID,
				Valid: true,
			},
			Amount: arg.Ammount,
		})
		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: sql.NullInt64{
				Int64: arg.FromAccountID,
				Valid: true,
			},
			Amount: -arg.Ammount,
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: sql.NullInt64{
				Int64: arg.ToAccountID,
				Valid: true,
			},
			Amount: arg.Ammount,
		})
		if err != nil {
			return err
		}

		// toAccount, err2 := q.GetAccountForUpdate(ctx, arg.ToAccountID)
		// if err2 != nil {
		// 	return err2
		// }

		result.ToAccount, err = q.AddToAccountBalance(ctx, AddToAccountBalanceParams{
			ID:     arg.ToAccountID,
			Amount: arg.Ammount,
		})
		if err != nil {
			return err
		}

		// fromAccount, err1 := q.GetAccountForUpdate(ctx, arg.FromAccountID)
		// if err1 != nil {
		// 	return err1
		// }

		result.FromAccount, err = q.AddToAccountBalance(ctx, AddToAccountBalanceParams{
			ID:     arg.FromAccountID,
			Amount: -arg.Ammount,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}
