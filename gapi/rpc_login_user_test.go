package gapi

import (
	"context"
	"testing"
	"time"

	mockdb "github.com/dubass83/simplebank/db/mock"
	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/dubass83/simplebank/pb"
	"github.com/dubass83/simplebank/util"
	mockwk "github.com/dubass83/simplebank/worker/mock"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestLoginUserGAPI(t *testing.T) {
	user, password := randomUser()
	sesionId, _ := uuid.FromBytes([]byte(password))
	refreshToken := util.RandomString(64)
	creeatedAt := time.Now()
	expiredAt := time.Now().Add(15 * time.Minute)

	testCases := []struct {
		name          string
		req           *pb.LoginUserRequest
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, resp *pb.LoginUserResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.LoginUserRequest{
				Username: user.Username,
				Password: password,
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetUser(gomock.Any(), user.Username).
					Times(1).
					Return(user, nil)

				// arg := &db.CreateSessionParams{
				// 	ID:           sesionId,
				// 	Username:     user.Username,
				// 	RefreshToken: refreshToken,
				// 	UserAgent:    "",
				// 	ClientIp:     "",
				// 	IsBloked:     false,
				// 	ExpiredAt:    expiredAt,
				// }
				store.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Session{
						ID:           sesionId,
						Username:     user.Username,
						RefreshToken: refreshToken,
						UserAgent:    "",
						ClientIp:     "",
						IsBloked:     false,
						ExpiredAt:    expiredAt,
						CreatedAt:    creeatedAt,
					}, nil)
			},
			checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				// createDbUser := res.GetUser()
				// require.Equal(t, user.Username, createDbUser.Username)
				// require.Equal(t, user.FullName, createDbUser.FullName)
				// require.Equal(t, user.Email, createDbUser.Email)

			},
		},
		// {
		// 	name: "InternalError",
		// 	req: &pb.LoginUserRequest{
		// 		Username: user.Username,
		// 		Password: password,
		// 		FullName: user.FullName,
		// 		Email:    user.Email,
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore, taskDistrebutor *mockwk.MockTaskDistributor) {
		// 		store.EXPECT().
		// 			LoginUserTx(gomock.Any(), gomock.Any()).
		// 			Times(1).
		// 			Return(db.LoginUserTxResult{}, sql.ErrConnDone)

		// 		taskDistrebutor.EXPECT().
		// 			DestributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		// 	checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
		// 		require.Error(t, err)
		// 		status, ok := status.FromError(err)
		// 		require.True(t, ok)
		// 		require.Equal(t, codes.Internal, status.Code())
		// 	},
		// }, {
		// 	name: "AlreadyExists",
		// 	req: &pb.LoginUserRequest{
		// 		Username: user.Username,
		// 		Password: password,
		// 		FullName: user.FullName,
		// 		Email:    user.Email,
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore, taskDistrebutor *mockwk.MockTaskDistributor) {
		// 		store.EXPECT().
		// 			LoginUserTx(gomock.Any(), gomock.Any()).
		// 			Times(1).
		// 			Return(db.LoginUserTxResult{}, db.ErrUniqueViolation)

		// 		taskDistrebutor.EXPECT().
		// 			DestributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		// 	checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
		// 		require.Error(t, err)
		// 		status, ok := status.FromError(err)
		// 		require.True(t, ok)
		// 		require.Equal(t, codes.AlreadyExists, status.Code())
		// 	},
		// }, {
		// 	name: "BadInputUsername",
		// 	req: &pb.LoginUserRequest{
		// 		Username: "mx",
		// 		Password: password,
		// 		FullName: user.FullName,
		// 		Email:    user.Email,
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore, taskDistrebutor *mockwk.MockTaskDistributor) {
		// 		store.EXPECT().
		// 			LoginUserTx(gomock.Any(), gomock.Any()).
		// 			Times(0)

		// 		taskDistrebutor.EXPECT().
		// 			DestributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		// 	checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
		// 		require.Error(t, err)
		// 		status, ok := status.FromError(err)
		// 		require.True(t, ok)
		// 		require.Equal(t, codes.InvalidArgument, status.Code())
		// 	},
		// }, {
		// 	name: "BadInputEmail",
		// 	req: &pb.LoginUserRequest{
		// 		Username: user.Username,
		// 		Password: password,
		// 		FullName: user.FullName,
		// 		Email:    "someATexampleDotCom",
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore, taskDistrebutor *mockwk.MockTaskDistributor) {
		// 		store.EXPECT().
		// 			LoginUserTx(gomock.Any(), gomock.Any()).
		// 			Times(0)

		// 		taskDistrebutor.EXPECT().
		// 			DestributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		// 	checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
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

			taskCtrl := gomock.NewController(t)
			defer taskCtrl.Finish()

			taskDistributor := mockwk.NewMockTaskDistributor(taskCtrl)

			// build stubs
			tc.buildStubs(store)

			// start test server and run gRPC function
			server := NewTestServer(t, store, taskDistributor)
			res, err := server.LoginUser(context.Background(), tc.req)

			tc.checkResponse(t, res, err)
		})

	}
}
