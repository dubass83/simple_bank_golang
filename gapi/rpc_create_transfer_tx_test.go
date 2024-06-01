package gapi

import (
	"context"
	"database/sql"
	"testing"
	"time"

	mockdb "github.com/dubass83/simplebank/db/mock"
	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/dubass83/simplebank/pb"
	"github.com/dubass83/simplebank/token"
	"github.com/dubass83/simplebank/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestCreateTransferTxGAPI(t *testing.T) {
	user1, _ := randomUser()
	user2, _ := randomUser()
	fromAccount := db.Account{
		ID:        1,
		Owner:     user1.Username,
		Balance:   100,
		Carrency:  util.UAH,
		CreatedAt: time.Now(),
	}
	toAccount := db.Account{
		ID:        2,
		Owner:     user2.Username,
		Balance:   200,
		Carrency:  util.UAH,
		CreatedAt: time.Now(),
	}
	// toAccountUSD := db.Account{
	// 	ID:        2,
	// 	Owner:     user2.Username,
	// 	Balance:   100,
	// 	Carrency:  util.USD,
	// 	CreatedAt: time.Now(),
	// }

	testCases := []struct {
		name          string
		req           *pb.CreateTransferTxRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, resp *pb.CreateTransferTxResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.CreateTransferTxRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId:   toAccount.ID,
				Amount:        10,
			},
			buildStubs: func(store *mockdb.MockStore) {
				frmAccId := fromAccount.ID
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(frmAccId)).
					Times(1).
					Return(fromAccount, nil)
				tAccId := toAccount.ID
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(tAccId)).
					Times(1).
					Return(toAccount, nil)
				arg := db.TransferTxParams{
					FromAccountID: fromAccount.ID,
					ToAccountID:   toAccount.ID,
					Ammount:       10,
				}
				transfer := db.Transfer{
					ID: 1,
					FromAccountID: sql.NullInt64{
						Int64: fromAccount.ID,
						Valid: true,
					},
					ToAccountID: sql.NullInt64{
						Int64: toAccount.ID,
						Valid: true,
					},
					Amount:    10,
					CreatedAt: time.Now(),
				}
				fromEntry := db.Entry{
					ID: 1,
					AccountID: sql.NullInt64{
						Int64: fromAccount.ID,
						Valid: true,
					},
					Amount:    -10,
					CreatedAt: time.Now(),
				}
				toEntry := db.Entry{
					ID: 1,
					AccountID: sql.NullInt64{
						Int64: toAccount.ID,
						Valid: true,
					},
					Amount:    10,
					CreatedAt: time.Now(),
				}
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.TransferTxResult{
						Transfer:    transfer,
						FromAccount: fromAccount,
						ToAccount:   toAccount,
						FromEntry:   fromEntry,
						ToEntry:     toEntry,
					}, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, user1.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferTxResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				// account := res.GetAccount()
				// require.Equal(t, user.Username, account.Owner)
				// require.Equal(t, initBalance, account.Balance)
				// require.Equal(t, currency, account.Carrency)
			},
		},
		// {
		// 	name: "InternalError",
		// 	req: &pb.CreateTransferTxRequest{
		// 		Currency: currency,
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		arg := db.CreateTransferTxParams{
		// 			Owner:    user.Username,
		// 			Balance:  initBalance,
		// 			Carrency: currency,
		// 		}
		// 		store.EXPECT().
		// 			CreateTransferTx(gomock.Any(), gomock.Eq(arg)).
		// 			Times(1).
		// 			Return(db.Account{}, sql.ErrConnDone)
		// 	},
		// 	buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
		// 		return BuildContext(t, tokenMaker, user.Username, time.Minute)
		// 	},
		// 	checkResponse: func(t *testing.T, res *pb.CreateTransferTxResponse, err error) {
		// 		require.Error(t, err)
		// 		require.Nil(t, res)
		// 		status, ok := status.FromError(err)
		// 		require.True(t, ok)
		// 		require.Equal(t, codes.Internal, status.Code())
		// 	},
		// }, {
		// 	name: "UnauthenticatedError",
		// 	req: &pb.CreateTransferTxRequest{
		// 		Currency: currency,
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().
		// 			CreateTransferTx(gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		// 	buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
		// 		return BuildContext(t, tokenMaker, user.Username, -time.Minute)
		// 	},
		// 	checkResponse: func(t *testing.T, res *pb.CreateTransferTxResponse, err error) {
		// 		require.Error(t, err)
		// 		require.Nil(t, res)
		// 		status, ok := status.FromError(err)
		// 		require.True(t, ok)
		// 		require.Equal(t, codes.Unauthenticated, status.Code())
		// 	},
		// }, {
		// 	name: "BadCurrency",
		// 	req: &pb.CreateTransferTxRequest{
		// 		Currency: badCurensy,
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().
		// 			CreateTransferTx(gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		// 	buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
		// 		return BuildContext(t, tokenMaker, user.Username, time.Minute)
		// 	},
		// 	checkResponse: func(t *testing.T, res *pb.CreateTransferTxResponse, err error) {
		// 		require.Error(t, err)
		// 		require.Nil(t, res)
		// 		status, ok := status.FromError(err)
		// 		require.True(t, ok)
		// 		require.Equal(t, codes.InvalidArgument, status.Code())
		// 	},
		// }, {
		// 	name: "AlreadyExistsError",
		// 	req: &pb.CreateTransferTxRequest{
		// 		Currency: currency,
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		arg := db.CreateTransferTxParams{
		// 			Owner:    user.Username,
		// 			Balance:  initBalance,
		// 			Carrency: currency,
		// 		}
		// 		store.EXPECT().
		// 			CreateTransferTx(gomock.Any(), gomock.Eq(arg)).
		// 			Times(1).
		// 			Return(db.Account{}, db.ErrUniqueViolation)
		// 	},
		// 	buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
		// 		return BuildContext(t, tokenMaker, user.Username, time.Minute)
		// 	},
		// 	checkResponse: func(t *testing.T, res *pb.CreateTransferTxResponse, err error) {
		// 		require.Error(t, err)
		// 		require.Nil(t, res)
		// 		status, ok := status.FromError(err)
		// 		require.True(t, ok)
		// 		require.Equal(t, codes.AlreadyExists, status.Code())
		// 	},
		// },
	}

	for i := range testCases {

		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			store := mockdb.NewMockStore(storeCtrl)
			// build stubs
			tc.buildStubs(store)
			// start test server and run gRPC function
			server := NewTestServer(t, store, nil)
			// create context
			ctx := tc.buildContext(t, server.tokenMaker)
			res, err := server.CreateTransfer(ctx, tc.req)
			// compare results
			tc.checkResponse(t, res, err)
		})

	}
}
