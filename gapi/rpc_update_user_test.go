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

func TestUpdateUserGAPI(t *testing.T) {
	user, _ := randomUser()
	newName := util.RandomOwner()
	newEmail := util.RandomEmail()

	testCases := []struct {
		name          string
		req           *pb.UpdateUserRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, resp *pb.UpdateUserResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateUserParams{
					Username: user.Username,
					FullName: sql.NullString{
						String: newName,
						Valid:  true,
					},
					Email: sql.NullString{
						String: newEmail,
						Valid:  true,
					},
				}
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.User{
						Username:          user.Username,
						HashedPassword:    user.HashedPassword,
						FullName:          newName,
						Email:             newEmail,
						PasswordChangedAt: user.PasswordChangedAt,
						CreatedAt:         user.PasswordChangedAt,
						IsEmailVerified:   user.IsEmailVerified,
					}, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				updateDbUser := res.GetUser()
				require.Equal(t, user.Username, updateDbUser.Username)
				require.Equal(t, newName, updateDbUser.FullName)
				require.Equal(t, newEmail, updateDbUser.Email)

			},
		},
		// {
		// 	name: "InternalError",
		// 	req: &pb.UpdateUserRequest{
		// 		Username: user.Username,
		// 		Password: password,
		// 		FullName: user.FullName,
		// 		Email:    user.Email,
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore, taskDistrebutor *mockwk.MockTaskDistributor) {
		// 		store.EXPECT().
		// 			UpdateUserTx(gomock.Any(), gomock.Any()).
		// 			Times(1).
		// 			Return(db.UpdateUserTxResult{}, sql.ErrConnDone)

		// 		taskDistrebutor.EXPECT().
		// 			DestributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		// 	checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
		// 		require.Error(t, err)
		// 		status, ok := status.FromError(err)
		// 		require.True(t, ok)
		// 		require.Equal(t, codes.Internal, status.Code())
		// 	},
		// },
		// {
		// 	name: "AlreadyExists",
		// 	req: &pb.UpdateUserRequest{
		// 		Username: user.Username,
		// 		Password: password,
		// 		FullName: user.FullName,
		// 		Email:    user.Email,
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore, taskDistrebutor *mockwk.MockTaskDistributor) {
		// 		store.EXPECT().
		// 			UpdateUserTx(gomock.Any(), gomock.Any()).
		// 			Times(1).
		// 			Return(db.UpdateUserTxResult{}, db.ErrUniqueViolation)

		// 		taskDistrebutor.EXPECT().
		// 			DestributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		// 	checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
		// 		require.Error(t, err)
		// 		status, ok := status.FromError(err)
		// 		require.True(t, ok)
		// 		require.Equal(t, codes.AlreadyExists, status.Code())
		// 	},
		// }, {
		// 	name: "BadInputUsername",
		// 	req: &pb.UpdateUserRequest{
		// 		Username: "mx",
		// 		Password: password,
		// 		FullName: user.FullName,
		// 		Email:    user.Email,
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore, taskDistrebutor *mockwk.MockTaskDistributor) {
		// 		store.EXPECT().
		// 			UpdateUserTx(gomock.Any(), gomock.Any()).
		// 			Times(0)

		// 		taskDistrebutor.EXPECT().
		// 			DestributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		// 	checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
		// 		require.Error(t, err)
		// 		status, ok := status.FromError(err)
		// 		require.True(t, ok)
		// 		require.Equal(t, codes.InvalidArgument, status.Code())
		// 	},
		// }, {
		// 	name: "BadInputEmail",
		// 	req: &pb.UpdateUserRequest{
		// 		Username: user.Username,
		// 		Password: password,
		// 		FullName: user.FullName,
		// 		Email:    "someATexampleDotCom",
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore, taskDistrebutor *mockwk.MockTaskDistributor) {
		// 		store.EXPECT().
		// 			UpdateUserTx(gomock.Any(), gomock.Any()).
		// 			Times(0)

		// 		taskDistrebutor.EXPECT().
		// 			DestributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		// 	checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
		// 		require.Error(t, err)
		// 		status, ok := status.FromError(err)
		// 		require.True(t, ok)
		// 		require.Equal(t, codes.InvalidArgument, status.Code())
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
			res, err := server.UpdateUser(ctx, tc.req)

			tc.checkResponse(t, res, err)
		})

	}
}
