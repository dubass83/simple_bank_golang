package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	n := 5
	ammount := int64(10)

	for i := 0; i < n; i++ {
		go func() {
			ctx := context.Background()
			result, err := testStore.TransferTx(ctx, TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Ammount:       ammount,
			})
			errs <- err
			results <- result

		}()
	}
	// check results
	existed := make(map[int]bool)
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
		result := <-results
		require.NotEmpty(t, result)

		// check transfer
		transfer := result.Transfer
		require.Equal(t, account1.ID, transfer.FromAccountID.Int64)
		require.Equal(t, account2.ID, transfer.ToAccountID.Int64)
		require.Equal(t, ammount, transfer.Amount)

		// check etries

		fromEntry := result.FromEntry
		require.Equal(t, account1.ID, fromEntry.AccountID.Int64)
		require.Equal(t, -ammount, fromEntry.Amount)

		toEntry := result.ToEntry
		require.Equal(t, account2.ID, toEntry.AccountID.Int64)
		require.Equal(t, ammount, toEntry.Amount)

		// check accounts balance
		fromAccount := result.FromAccount
		require.Equal(t, account1.ID, fromAccount.ID)
		require.Equal(t, account1.Carrency, fromAccount.Carrency)
		require.Equal(t, account1.Owner, fromAccount.Owner)
		require.NotEmpty(t, fromAccount.CreatedAt)

		toAccount := result.ToAccount
		require.Equal(t, account2.ID, toAccount.ID)
		require.Equal(t, account2.Carrency, toAccount.Carrency)
		require.Equal(t, account2.Owner, toAccount.Owner)
		require.NotEmpty(t, toAccount.CreatedAt)

		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%ammount == 0)

		k := int(diff1 / ammount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	updatedAccount1, err := testStore.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	updatedAccount2, err := testStore.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance-int64(n)*ammount, updatedAccount1.Balance)
	require.Equal(t, account2.Balance+int64(n)*ammount, updatedAccount2.Balance)
}

func TestDeadlockInTransferTx(t *testing.T) {

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	errs := make(chan error)

	n := 10
	ammount := int64(10)

	for i := 0; i < n; i++ {
		fromAccountID := account1.ID
		toAccountID := account2.ID
		if i%2 == 0 {
			fromAccountID = account2.ID
			toAccountID = account1.ID
		}
		go func() {
			ctx := context.Background()
			_, err := testStore.TransferTx(ctx, TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Ammount:       ammount,
			})
			errs <- err

		}()
	}
	// check results
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	updatedAccount1, err := testStore.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	updatedAccount2, err := testStore.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance, updatedAccount1.Balance)
	require.Equal(t, account2.Balance, updatedAccount2.Balance)
}
