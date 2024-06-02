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
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetUser(t *testing.T) {
	user, _ := randomUser()
	badUser := "dubass83"
	testCases := []struct {
		name          string
		req           *pb.GetUserRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponce func(t *testing.T, res *pb.GetUserResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.GetUserRequest{
				Username: user.Username,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := user.Username
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(user, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponce: func(t *testing.T, res *pb.GetUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Equal(t, user.Username, res.User.Username)
				require.Equal(t, user.Email, res.User.Email)
				require.Equal(t, user.FullName, res.User.FullName)
			},
		}, {
			name: "InternalError",
			req: &pb.GetUserRequest{
				Username: user.Username,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := user.Username
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponce: func(t *testing.T, res *pb.GetUserResponse, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				status, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, status.Code())
			},
		}, {
			name: "UnauthenticatedError",
			req: &pb.GetUserRequest{
				Username: user.Username,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, user.Username, -time.Minute)
			},
			checkResponce: func(t *testing.T, res *pb.GetUserResponse, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				status, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, status.Code())
			},
		}, {
			name: "UnauthenticatedUserError",
			req: &pb.GetUserRequest{
				Username: user.Username,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, badUser, time.Minute)
			},
			checkResponce: func(t *testing.T, res *pb.GetUserResponse, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				status, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, status.Code())
				errStr := fmt.Sprintf("rpc error: code = Unauthenticated desc = unauthorized: user: %s is not authorized to get information about another user: %s", badUser, user.Username)
				require.EqualError(t,
					err,
					errStr)
			},
		}, {
			name: "NotFoundError",
			req: &pb.GetUserRequest{
				Username: user.Username,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := user.Username
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponce: func(t *testing.T, res *pb.GetUserResponse, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				status, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, status.Code())
			},
		}, {
			name: "BadInputError",
			req: &pb.GetUserRequest{
				Username: "o1",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return BuildContext(t, tokenMaker, badUser, time.Minute)
			},
			checkResponce: func(t *testing.T, res *pb.GetUserResponse, err error) {
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
			control := gomock.NewController(t)
			defer control.Finish()
			store := mockdb.NewMockStore(control)
			tc.buildStubs(store)
			server := NewTestServer(t, store, nil)
			ctx := tc.buildContext(t, server.tokenMaker)
			res, err := server.GetUser(ctx, tc.req)
			tc.checkResponce(t, res, err)
		})
	}
}
