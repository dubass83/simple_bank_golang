package db

import (
	"context"
	"testing"

	"github.com/dubass83/simplebank/util"
	"github.com/stretchr/testify/require"
)

func TestCreateAccount(t *testing.T) {
	arg := CreateAccountParams{
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Carrency: util.RandomCurrency(),
	}
	account, err := TestQueries.CreateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.Balance, account.Balance)
	require.Equal(t, arg.Carrency, account.Carrency)

	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)
}
