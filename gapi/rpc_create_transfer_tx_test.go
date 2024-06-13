package gapi

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	mockdb "github.com/dubass83/simplebank/db/mock"
	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/dubass83/simplebank/pb"
	"github.com/dubass83/simplebank/token"
	"github.com/dubass83/simplebank/util"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	toAccountUSD := db.Account{
		ID:        2,
		Owner:     user2.Username,
		Balance:   100,
		Carrency:  util.USD,
		CreatedAt: time.Now(),
	}

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
					FromAccountID: pgtype.Int8{
						Int64: fromAccount.ID,
						Valid: true,
					},
					ToAccountID: pgtype.Int8{
						Int64: toAccount.ID,
						Valid: true,
					},
					Amount:    10,
					CreatedAt: time.Now(),
				}
				fromEntry := db.Entry{
					ID: 1,
					AccountID: pgtype.Int8{
						Int64: fromAccount.ID,
						Valid: true,
					},
					Amount:    -10,
					CreatedAt: time.Now(),
				}
				toEntry := db.Entry{
					ID: 1,
					AccountID: pgtype.Int8{
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
				return BuildContext(t, tokenMaker, user1.Username, user1.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferTxResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
			},
		}, {
			name: "Account1NotFound",
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
					Return(db.Account{}, fmt.Errorf("account not found: %d", frmAccId))
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, user1.Username, user1.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferTxResponse, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				require.EqualError(t, err, "cannot get Account: account not found: 1")
			},
		}, {
			name: "Account2NotFound",
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
					Return(db.Account{}, fmt.Errorf("account not found: %d", tAccId))
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, user1.Username, user1.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferTxResponse, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				require.EqualError(t, err, "cannot get Account: account not found: 2")
			},
		}, {
			name: "InternalError",
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
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.TransferTxResult{}, sql.ErrConnDone)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, user1.Username, user1.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferTxResponse, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				status, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, status.Code())
			},
		}, {
			name: "UnauthenticatedError",
			req: &pb.CreateTransferTxRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId:   toAccount.ID,
				Amount:        10,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, user1.Username, user1.Role, -time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferTxResponse, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				status, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, status.Code())
			},
		}, {
			name: "NotUserToken",
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

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, user2.Username, user2.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferTxResponse, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				require.EqualError(t, err, "user allowed to transfer money only from his account")
			},
		}, {
			name: "DifferentCurrencyError",
			req: &pb.CreateTransferTxRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId:   toAccountUSD.ID,
				Amount:        10,
			},
			buildStubs: func(store *mockdb.MockStore) {
				frmAccId := fromAccount.ID
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(frmAccId)).
					Times(1).
					Return(fromAccount, nil)
				tAccId := toAccountUSD.ID
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(tAccId)).
					Times(1).
					Return(toAccountUSD, nil)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, user1.Username, user1.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferTxResponse, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				status, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, status.Code())
			},
		},
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
