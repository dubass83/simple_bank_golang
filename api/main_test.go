package api

import (
	"os"
	"testing"
	"time"

	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/dubass83/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func NewTestServer(t *testing.T, store db.Store) *Server {
	config := util.Config{
		TokenString:   util.RandomString(32),
		TokenDuration: time.Minute * 5,
	}
	server, err := NewServer(config, store)
	require.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
