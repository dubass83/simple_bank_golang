package db

import (
	"context"
	"testing"
	"time"

	"github.com/dubass83/simplebank/util"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	password := util.RandomString(8)
	hashPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	arg := CreateUserParams{
		Username:       util.RandomOwner(),
		HashedPassword: hashPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
	user, err := TestQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)

	require.NotZero(t, user.PasswordChangedAt)
	require.NotZero(t, user.CreatedAt)
	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)
	user2, err := TestQueries.GetUser(context.Background(), user1.Username)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.HashedPassword, user2.HashedPassword)
	require.Equal(t, user1.FullName, user2.FullName)
	require.Equal(t, user1.Email, user2.Email)
	require.WithinDuration(t, user1.PasswordChangedAt, user2.PasswordChangedAt, time.Second)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
}

func TestUpdateUserFullName(t *testing.T) {
	userOld := createRandomUser(t)
	newFullName := util.RandomOwner()

	userUpdated, err := TestQueries.UpdateUser(context.Background(), UpdateUserParams{
		Username: userOld.Username,
		FullName: pgtype.Text{
			String: newFullName,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEqual(t, userOld.FullName, userUpdated.FullName)
	require.Equal(t, newFullName, userUpdated.FullName)
}

func TestUpdateUserEmail(t *testing.T) {
	userOld := createRandomUser(t)
	newEmail := util.RandomEmail()

	userUpdated, err := TestQueries.UpdateUser(context.Background(), UpdateUserParams{
		Username: userOld.Username,
		Email: pgtype.Text{
			String: newEmail,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEqual(t, userOld.Email, userUpdated.Email)
	require.Equal(t, newEmail, userUpdated.Email)
}

func TestUpdateUserHashedPassword(t *testing.T) {
	userOld := createRandomUser(t)
	hp, err := util.HashPassword(util.RandomString(8))
	require.NoError(t, err)

	userUpdated, err := TestQueries.UpdateUser(context.Background(), UpdateUserParams{
		Username: userOld.Username,
		HashedPassword: pgtype.Text{
			String: hp,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEqual(t, userOld.HashedPassword, userUpdated.HashedPassword)
	require.Equal(t, hp, userUpdated.HashedPassword)
}

func TestUpdateUserHashedAll(t *testing.T) {
	userOld := createRandomUser(t)
	newEmail := util.RandomEmail()
	newFullName := util.RandomOwner()
	hp, err := util.HashPassword(util.RandomString(8))
	require.NoError(t, err)

	userUpdated, err := TestQueries.UpdateUser(context.Background(), UpdateUserParams{
		Username: userOld.Username,
		HashedPassword: pgtype.Text{
			String: hp,
			Valid:  true,
		},
		Email: pgtype.Text{
			String: newEmail,
			Valid:  true,
		},
		FullName: pgtype.Text{
			String: newFullName,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEqual(t, userOld.HashedPassword, userUpdated.HashedPassword)
	require.NotEqual(t, userOld.Email, userUpdated.Email)
	require.NotEqual(t, userOld.FullName, userUpdated.FullName)
	require.Equal(t, newEmail, userUpdated.Email)
	require.Equal(t, newFullName, userUpdated.FullName)
	require.Equal(t, hp, userUpdated.HashedPassword)
}
