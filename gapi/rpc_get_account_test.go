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

func TestGetAccount(t *testing.T) {
	user, _ := randomUser()
	account := db.Account{
		ID:        1,
		Owner:     user.Username,
		Balance:   0,
		Carrency:  util.UAH,
		CreatedAt: time.Now(),
	}

	testCases := []struct {
		name          string
		req           *pb.GetAccountRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponce func(t *testing.T, res *pb.GetAccountResponse, err error)
	}{
		{
			name: "OK",
			req:  &pb.GetAccountRequest{Id: account.ID},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponce: func(t *testing.T, res *pb.GetAccountResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Equal(t, account.ID, res.Account.Id)
				require.Equal(t, account.Owner, res.Account.Owner)
				require.Equal(t, account.Balance, res.Account.Balance)
				require.Equal(t, account.Carrency, res.Account.Carrency)
			},
		}, {
			name: "InternalError",
			req:  &pb.GetAccountRequest{Id: account.ID},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponce: func(t *testing.T, res *pb.GetAccountResponse, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				status, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, status.Code())
			},
		}, {
			name: "NotFoundError",
			req:  &pb.GetAccountRequest{Id: account.ID},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, db.ErrRecordNotFound)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponce: func(t *testing.T, res *pb.GetAccountResponse, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				status, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, status.Code())
			},
		}, {
			name: "BadUserError",
			req:  &pb.GetAccountRequest{Id: account.ID},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, "dubass83", time.Minute)
			},
			checkResponce: func(t *testing.T, res *pb.GetAccountResponse, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				status, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, status.Code())
			},
		}, {
			name: "BadTokenError",
			req:  &pb.GetAccountRequest{Id: account.ID},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, user.Username, -time.Minute)
			},
			checkResponce: func(t *testing.T, res *pb.GetAccountResponse, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				status, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, status.Code())
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)
			server := NewTestServer(t, store, nil)
			ctx := tc.buildContext(t, server.tokenMaker)
			res, err := server.GetAccount(ctx, tc.req)
			tc.checkResponce(t, res, err)
		})
	}
}
