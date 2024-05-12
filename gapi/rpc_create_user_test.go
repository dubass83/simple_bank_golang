package gapi

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	mockdb "github.com/dubass83/simplebank/db/mock"
	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/dubass83/simplebank/pb"
	"github.com/dubass83/simplebank/util"
	mockwk "github.com/dubass83/simplebank/worker/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type eqCreateUserTxParamMatcher struct {
	arg      db.CreateUserTxParams
	password string
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

	return reflect.DeepEqual(expected.arg.CreateUserParams, actualArg.CreateUserParams)
}

func (e eqCreateUserTxParamMatcher) String() string {
	return fmt.Sprintf("is argument %v and password %s", e.arg, e.password)
}

func eqCreateUserTxParam(arg db.CreateUserTxParams, pass string) gomock.Matcher {
	return eqCreateUserTxParamMatcher{arg, pass}
}

func TestCreateUserGAPI(t *testing.T) {
	user, password := randomUser()

	testCases := []struct {
		name          string
		req           *pb.CreateUserRequest
		buildStubs    func(store *mockdb.MockStore)
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
			buildStubs: func(store *mockdb.MockStore) {
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
					CreateUserTx(gomock.Any(), eqCreateUserTxParam(arg, password)).
					Times(1).
					Return(db.CreateUserTxResult{User: user}, nil)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				createDbUser := res.GetUser()
				require.Equal(t, user.Username, createDbUser.Username)
				require.Equal(t, user.FullName, createDbUser.FullName)
				require.Equal(t, user.Email, createDbUser.Email)

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
