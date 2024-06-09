package gapi

import (
	"context"
	"errors"
	"fmt"

	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/dubass83/simplebank/pb"
	"github.com/dubass83/simplebank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (srv *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	payload, err := srv.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if violations := validateGetUserRequest(req); violations != nil {
		return nil, invalidArgumentError(violations)
	}
	if payload.Username != req.GetUsername() {
		err := fmt.Errorf("user: %s is not authorized to get information about another user: %s",
			payload.Username,
			req.GetUsername(),
		)
		return nil, unauthenticatedError(err)
	}
	user, err := srv.store.GetUser(ctx, req.GetUsername())
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "cannot get user: %s", err)
	}

	rsp := &pb.GetUserResponse{
		User: convertUser(user),
	}
	return rsp, nil
}

func validateGetUserRequest(req *pb.GetUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	return
}
