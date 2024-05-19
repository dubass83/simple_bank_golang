package gapi

import (
	"context"
	"database/sql"
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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
				require.Equal(t, user.Username, res.User.Username)
				require.Equal(t, user.FullName, res.User.FullName)
				require.Equal(t, sesionId.String(), res.SessionId)
			},
		}, {
			name: "InternalErrorGetUser",
			req: &pb.LoginUserRequest{
				Username: user.Username,
				Password: password,
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetUser(gomock.Any(), user.Username).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)

				store.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, st.Code(), codes.Internal)
			},
		}, {
			name: "InternalErrorCreateSession",
			req: &pb.LoginUserRequest{
				Username: user.Username,
				Password: password,
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetUser(gomock.Any(), user.Username).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Session{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, st.Code(), codes.Internal)
			},
		}, {
			name: "BadInputUsername",
			req: &pb.LoginUserRequest{
				Username: "mx",
				Password: password,
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetUser(gomock.Any(), user.Username).
					Times(0)

				store.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, st.Code(), codes.InvalidArgument)
			},
		},
		{
			name: "BadInputPassword",
			req: &pb.LoginUserRequest{
				Username: user.Username,
				Password: "",
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetUser(gomock.Any(), user.Username).
					Times(0)

				store.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, st.Code(), codes.InvalidArgument)
			},
		},
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
