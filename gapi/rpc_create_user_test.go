package gapi

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"testing"

	mockdb "github.com/dubass83/simplebank/db/mock"
	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/dubass83/simplebank/pb"
	"github.com/dubass83/simplebank/util"
	"github.com/dubass83/simplebank/worker"
	mockwk "github.com/dubass83/simplebank/worker/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type eqCreateUserTxParamMatcher struct {
	arg      db.CreateUserTxParams
	password string
	user     db.User
}

func (expected eqCreateUserTxParamMatcher) Matches(x any) bool {
	actualArg, ok := x.(db.CreateUserTxParams)
	if !ok {
		return false
	}
	err := util.CheckPassword(expected.password, actualArg.HashedPassword)
	if err != nil {
		return false
	}
	expected.arg.HashedPassword = actualArg.HashedPassword

	if !reflect.DeepEqual(expected.arg.CreateUserParams, actualArg.CreateUserParams) {
		return false
	}
	err = actualArg.AfterFunc(expected.user)
	return err == nil
}

func (e eqCreateUserTxParamMatcher) String() string {
	return fmt.Sprintf("is argument %v and password %s", e.arg, e.password)
}

func eqCreateUserTxParam(arg db.CreateUserTxParams, pass string, user db.User) gomock.Matcher {
	return eqCreateUserTxParamMatcher{arg, pass, user}
}

func TestCreateUserGAPI(t *testing.T) {
	user, password := randomUser()

	testCases := []struct {
		name          string
		req           *pb.CreateUserRequest
		buildStubs    func(store *mockdb.MockStore, taskDistrebutor *mockwk.MockTaskDistributor)
		checkResponse func(t *testing.T, resp *pb.CreateUserResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistrebutor *mockwk.MockTaskDistributor) {
				arg := db.CreateUserTxParams{
					CreateUserParams: db.CreateUserParams{
						Username: user.Username,
						FullName: user.FullName,
						Email:    user.Email,
					},
					// AfterFunc: func(user db.User) error {
					// },
				}
				store.EXPECT().
					CreateUserTx(gomock.Any(), eqCreateUserTxParam(arg, password, user)).
					Times(1).
					Return(db.CreateUserTxResult{User: user}, nil)

				payload := &worker.PayloadSendVerifyEmail{
					Username: user.Username,
				}
				taskDistrebutor.EXPECT().
					DestributeTaskSendVerifyEmail(gomock.Any(), payload, gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				createDbUser := res.GetUser()
				require.Equal(t, user.Username, createDbUser.Username)
				require.Equal(t, user.FullName, createDbUser.FullName)
				require.Equal(t, user.Email, createDbUser.Email)

			},
		}, {
			name: "InternalError",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistrebutor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.CreateUserTxResult{}, sql.ErrConnDone)

				taskDistrebutor.EXPECT().
					DestributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				status, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, status.Code())
			},
		}, {
			name: "AlreadyExists",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistrebutor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.CreateUserTxResult{}, db.ErrUniqueViolation)

				taskDistrebutor.EXPECT().
					DestributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				status, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.AlreadyExists, status.Code())
			},
		}, {
			name: "BadInputUsername",
			req: &pb.CreateUserRequest{
				Username: "mx",
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistrebutor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(0)

				taskDistrebutor.EXPECT().
					DestributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				status, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, status.Code())
			},
		}, {
			name: "BadInputEmail",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    "someATexampleDotCom",
			},
			buildStubs: func(store *mockdb.MockStore, taskDistrebutor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(0)

				taskDistrebutor.EXPECT().
					DestributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
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

			taskCtrl := gomock.NewController(t)
			defer taskCtrl.Finish()

			taskDistributor := mockwk.NewMockTaskDistributor(taskCtrl)

			// build stubs
			tc.buildStubs(store, taskDistributor)

			// start test server and run gRPC function
			server := NewTestServer(t, store, taskDistributor)
			res, err := server.CreateUser(context.Background(), tc.req)

			tc.checkResponse(t, res, err)
		})

	}
}

func randomUser() (db.User, string) {
	password := util.RandomString(8)
	hash, err := util.HashPassword(password)
	if err != nil {
		return db.User{}, ""
	}

	user := db.User{
		Username:       util.RandomOwner(),
		HashedPassword: hash,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}

	return user, password
}
