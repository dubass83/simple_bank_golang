package gapi

import (
	"context"
	"fmt"
	"testing"
	"time"

	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/dubass83/simplebank/token"
	"github.com/dubass83/simplebank/util"
	"github.com/dubass83/simplebank/worker"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
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

func BuildContext(t *testing.T, tokenMaker token.Maker, username string, role string, duration time.Duration) context.Context {
	ctx := context.Background()
	token, _, err := tokenMaker.CreateToken(username, role, duration)
	require.NoError(t, err)
	barierToken := fmt.Sprintf("%s %s", authType, token)
	md := metadata.MD{
		authorizationHeader: []string{
			barierToken,
		},
	}
	return metadata.NewIncomingContext(ctx, md)
}
