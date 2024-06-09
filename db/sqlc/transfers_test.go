package db

import (
	"context"
	"testing"

	"github.com/dubass83/simplebank/util"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func createRandomTransfer(t *testing.T) Transfer {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	arg := CreateTransferParams{
		FromAccountID: pgtype.Int8{
			Int64: account1.ID,
			Valid: true,
		},
		ToAccountID: pgtype.Int8{
			Int64: account2.ID,
			Valid: true,
		},
		Amount: util.RandomMoney(),
	}

	transfer, err := TestQueries.CreateTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	require.Equal(t, account1.ID, transfer.FromAccountID.Int64)
	require.Equal(t, account2.ID, transfer.ToAccountID.Int64)
	require.Equal(t, arg.Amount, transfer.Amount)

	return transfer
}

func TestCreateTransfer(t *testing.T) {
	createRandomTransfer(t)
}

func TestGetTransfer(t *testing.T) {
	transfer1 := createRandomTransfer(t)

	transfer2, err := TestQueries.GetTransfer(context.Background(), transfer1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, transfer2)

	require.Equal(t, transfer1.ID, transfer2.ID)
	require.Equal(t, transfer1.FromAccountID, transfer2.FromAccountID)
	require.Equal(t, transfer1.ToAccountID, transfer2.ToAccountID)
	require.Equal(t, transfer1.Amount, transfer2.Amount)
}

func TestUpdateTransfer(t *testing.T) {
	transfer1 := createRandomTransfer(t)

	arg := UpdateTransferParams{
		ID:     transfer1.ID,
		Amount: util.RandomMoney(),
	}

	transfer2, err := TestQueries.UpdateTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfer2)

	require.Equal(t, transfer1.FromAccountID, transfer2.FromAccountID)
	require.Equal(t, transfer1.ToAccountID, transfer2.ToAccountID)
	require.Equal(t, arg.Amount, transfer2.Amount)
}

func TestDeleteTransfer(t *testing.T) {
	transfer1 := createRandomTransfer(t)

	err := TestQueries.DeleteTransfer(context.Background(), transfer1.ID)
	require.NoError(t, err)

	transfer2, err := TestQueries.GetTransfer(context.Background(), transfer1.ID)
	require.Error(t, err)
	require.Empty(t, transfer2)
}

func TestListTransfers(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomTransfer(t)
	}

	arg := ListTransfersParams{
		Limit:  5,
		Offset: 5,
	}

	transfers, err := TestQueries.ListTransfers(context.Background(), arg)
	require.NoError(t, err)

	for _, transfer := range transfers {
		require.NotEmpty(t, transfer)
	}
}
