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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateAccountGAPI(t *testing.T) {
	user, _ := randomUser()
	currency := util.UAH
	initBalance := int64(0)
	badCurensy := "PLN"

	testCases := []struct {
		name          string
		req           *pb.CreateAccountRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, resp *pb.CreateAccountResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.CreateAccountRequest{
				Currency: currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateAccountParams{
					Owner:    user.Username,
					Balance:  initBalance,
					Carrency: currency,
				}
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Account{
						ID:        1,
						Owner:     user.Username,
						Balance:   initBalance,
						Carrency:  currency,
						CreatedAt: time.Now(),
					}, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, user.Username, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateAccountResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				account := res.GetAccount()
				require.Equal(t, user.Username, account.Owner)
				require.Equal(t, initBalance, account.Balance)
				require.Equal(t, currency, account.Carrency)

			},
		}, {
			name: "InternalError",
			req: &pb.CreateAccountRequest{
				Currency: currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateAccountParams{
					Owner:    user.Username,
					Balance:  initBalance,
					Carrency: currency,
				}
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, user.Username, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateAccountResponse, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				status, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, status.Code())
			},
		}, {
			name: "UnauthenticatedError",
			req: &pb.CreateAccountRequest{
				Currency: currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, user.Username, user.Role, -time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateAccountResponse, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				status, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, status.Code())
			},
		}, {
			name: "BadCurrency",
			req: &pb.CreateAccountRequest{
				Currency: badCurensy,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, user.Username, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateAccountResponse, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				status, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, status.Code())
			},
		}, {
			name: "AlreadyExistsError",
			req: &pb.CreateAccountRequest{
				Currency: currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateAccountParams{
					Owner:    user.Username,
					Balance:  initBalance,
					Carrency: currency,
				}
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Account{}, db.ErrUniqueViolation)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, user.Username, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateAccountResponse, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				status, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.AlreadyExists, status.Code())
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
			res, err := server.CreateAccount(ctx, tc.req)
			// compare results
			tc.checkResponse(t, res, err)
		})

	}
}
