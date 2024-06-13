package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	mockdb "github.com/dubass83/simplebank/db/mock"
	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/dubass83/simplebank/token"
	"github.com/dubass83/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type eqCreateUserParamMatcher struct {
	arg      db.CreateUserParams
	password string
}

func (e eqCreateUserParamMatcher) Matches(x any) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}
	err := util.CheckPassword(e.password, arg.HashedPassword)
	if err != nil {
		return false
	}
	e.arg.HashedPassword = arg.HashedPassword

	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateUserParamMatcher) String() string {
	return fmt.Sprintf("is argument %v and password %s", e.arg, e.password)
}

func eqCreateUserParam(arg db.CreateUserParams, pass string) gomock.Matcher {
	return eqCreateUserParamMatcher{arg, pass}
}

func TestGetUserAPI(t *testing.T) {
	user, _ := randomUser()

	testCases := []struct {
		name          string
		username      string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "OK",
			username: user.Username,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				AddAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, user.Role, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, userResponse{
					Username: user.Username,
					FullName: user.FullName,
					Email:    user.Email,
				})
			},
		},
		{
			name:     "NotFound",
			username: "notFound",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				AddAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, user.Role, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq("notFound")).
					Times(1).
					Return(db.User{}, db.ErrRecordNotFound)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:     "BadRequest",
			username: "b@du$#r",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				AddAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, user.Role, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:     "InternalServerError",
			username: user.Username,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				AddAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, user.Role, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {

		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			// build stubs
			tc.buildStubs(store)

			// start test server and send request
			server := NewTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/users/%s", tc.username)

			reqest, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, reqest, server.tokenMaker)

			server.router.ServeHTTP(recorder, reqest)
			// check return status code
			tc.checkResponse(t, recorder)
		})

	}

}

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser()

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserParams{
					Username: user.Username,
					FullName: user.FullName,
					Email:    user.Email,
				}
				store.EXPECT().
					CreateUser(gomock.Any(), eqCreateUserParam(arg, password)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, userResponse{
					Username: user.Username,
					FullName: user.FullName,
					Email:    user.Email,
				})
			},
		},
		{
			name: "BadRequest",
			body: gin.H{
				"username":  "b@dU$#r",
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "ForbidenError",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, db.ErrUniqueViolation)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {

		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			// build stubs
			tc.buildStubs(store)

			// start test server and send request
			server := NewTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := "/users"
			// convert map to json
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			reqest, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, reqest)
			// check return status code
			tc.checkResponse(t, recorder)
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
		Role:           util.DepositorRole,
		HashedPassword: hash,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}

	return user, password
}

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, uResp userResponse) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	var gotUserResponse userResponse
	err = json.Unmarshal(data, &gotUserResponse)
	require.NoError(t, err)

	require.Equal(t, gotUserResponse, uResp)
}
