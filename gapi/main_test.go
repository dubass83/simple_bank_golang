package gapi

import (
	"testing"
	"time"

	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/dubass83/simplebank/util"
	"github.com/dubass83/simplebank/worker"
	"github.com/stretchr/testify/require"
)

func NewTestServer(t *testing.T, store db.Store, taskDistributor worker.TaskDistributor) *Server {
	config := util.Config{
		TokenString:   util.RandomString(32),
		TokenDuration: time.Minute * 5,
	}
	server, err := NewServer(config, store, taskDistributor)
	require.NoError(t, err)

	return server
}
