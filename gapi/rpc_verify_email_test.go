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
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestVerifyEmailGAPI(t *testing.T) {
	user, _ := randomUser()
	id := util.RandomInt(1, 100)
	secretCode := util.RandomString(32)

	testCases := []struct {
		name          string
		req           *pb.VerifyEmailRequest
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, resp *pb.VerifyEmailResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.VerifyEmailRequest{
				Id:         id,
				SecretCode: secretCode,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.VerifyEmailTxParams{
					ID:         id,
					SecretCode: secretCode,
				}
				user = db.User{
					Username:          user.Username,
					HashedPassword:    user.HashedPassword,
					FullName:          user.FullName,
					Email:             user.Email,
					PasswordChangedAt: user.PasswordChangedAt,
					CreatedAt:         user.CreatedAt,
					IsEmailVerified:   true,
				}
				verifyEmail := db.VerifyEmail{
					ID:         id,
					Username:   user.Username,
					Email:      user.Email,
					SecretCode: secretCode,
					IsUsed:     true,
					CreatedAt:  time.Now(),
					ExpiredAt:  time.Now().Add(time.Minute),
				}
				res := db.VerifyEmailTxResult{
					User:        user,
					VerifyEmail: verifyEmail,
				}
				store.EXPECT().
					VerifyEmailTx(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(res, nil)
			},
			checkResponse: func(t *testing.T, res *pb.VerifyEmailResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.True(t, res.GetIsVerified())
			},
		}, {
			name: "InternalError",
			req: &pb.VerifyEmailRequest{
				Id:         id,
				SecretCode: secretCode,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.VerifyEmailTxParams{
					ID:         id,
					SecretCode: secretCode,
				}
				store.EXPECT().
					VerifyEmailTx(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.VerifyEmailTxResult{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, res *pb.VerifyEmailResponse, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, st.Code(), codes.Internal)
			},
		}, {
			name: "BadInputID",
			req: &pb.VerifyEmailRequest{
				Id:         -1,
				SecretCode: secretCode,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					VerifyEmailTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.VerifyEmailResponse, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, st.Code(), codes.InvalidArgument)
			},
		}, {
			name: "BadInputSecretCode",
			req: &pb.VerifyEmailRequest{
				Id:         id,
				SecretCode: "qwerty",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					VerifyEmailTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.VerifyEmailResponse, err error) {
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

			// build stubs
			tc.buildStubs(store)

			// start test server and run gRPC function
			server := NewTestServer(t, store, nil)
			res, err := server.VerifyEmail(context.Background(), tc.req)

			tc.checkResponse(t, res, err)
		})

	}
}
