package gapi

import (
	"context"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

func GrpcLogger(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	log.Info().
		Str("func", "GrpcLogger").
		Msg("receive GRPC request")
	result, err := handler(ctx, req)
	return result, err
}
