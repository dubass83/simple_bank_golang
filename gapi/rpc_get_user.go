package gapi

import (
	"context"
	"database/sql"

	"github.com/dubass83/simplebank/pb"
	"github.com/dubass83/simplebank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (srv *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	if violations := validateGetUserRequest(req); violations != nil {
		return nil, invalidArgumentError(violations)
	}
	user, err := srv.store.GetUser(ctx, req.GetUsername())
	if err != nil {
		if err == sql.ErrNoRows {
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
	if err := val.ValidateUsername(req.Username); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	return
}
