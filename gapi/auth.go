package gapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/dubass83/simplebank/token"
	"google.golang.org/grpc/metadata"
)

const (
	authorizationHeader = "authorization"
	authType            = "bearer"
)

func (server *Server) authorizeUser(ctx context.Context) (*token.Payload, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("cannot get metadata from incoming context")
	}
	values := md.Get(authorizationHeader)
	if len(values) == 0 {
		return nil, fmt.Errorf("authorization token does not provided in the metadata")
	}
	authHeader := values[0]
	fields := strings.Split(authHeader, " ")
	if authType != strings.ToLower(fields[0]) {
		return nil, fmt.Errorf("not supported authorization type: %s", strings.ToLower(fields[0]))
	}
	payload, err := server.tokenMaker.VerifyToken(fields[1])
	if err != nil {
		return nil, fmt.Errorf("invalid token: %s", err)
	}
	return payload, nil
}
