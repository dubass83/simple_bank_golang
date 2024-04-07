package gapi

import (
	"fmt"

	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/dubass83/simplebank/pb"
	"github.com/dubass83/simplebank/token"
	"github.com/dubass83/simplebank/util"
	"github.com/dubass83/simplebank/worker"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
	config          util.Config
	store           db.Store
	tokenMaker      token.Maker
	taskDestributor worker.TaskDistributor
}

// NewServer creates a new gRPC server
func NewServer(config util.Config, store db.Store, taskDestributor worker.TaskDistributor) (*Server, error) {
	tokenMaker, err := token.NewJwtMaker(config.TokenString)
	if err != nil {
		return nil, fmt.Errorf("can not create token maker: %w", err)
	}
	server := &Server{
		config:          config,
		store:           store,
		tokenMaker:      tokenMaker,
		taskDestributor: taskDestributor,
	}

	return server, nil
}
