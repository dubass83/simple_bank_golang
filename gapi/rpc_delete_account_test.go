package gapi

import (
	"context"
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

func TestDeleteAccount(t *testing.T) {
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
		req           *pb.DeleteAccountRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponce func(t *testing.T, res *pb.DeleteAccountResponse, err error)
	}{
		{
			name: "OK",
			req:  &pb.DeleteAccountRequest{Id: account.ID},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)

				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponce: func(t *testing.T, res *pb.DeleteAccountResponse, err error) {
				require.NotNil(t, res)
				require.NoError(t, err)
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
			res, err := server.DeleteAccount(ctx, tc.req)
			tc.checkResponce(t, res, err)
		})
	}
}
