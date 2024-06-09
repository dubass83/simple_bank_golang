package gapi

import (
	"context"
	"errors"

	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/dubass83/simplebank/pb"
	"github.com/dubass83/simplebank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (srv *Server) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	payload, err := srv.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if violations := validateGetAccountRequest(req); violations != nil {
		return nil, invalidArgumentError(violations)
	}

	Account, err := srv.store.GetAccount(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "account not found: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "cannot get Account: %s", err)
	}

	if payload.Username != Account.Owner {
		return nil, status.Errorf(
			codes.Unauthenticated,
			"user: %s not allouwd to get info for account ID: %d",
			payload.Username,
			req.GetId())
	}

	rsp := &pb.GetAccountResponse{
		Account: convertAccount(Account),
	}
	return rsp, nil
}

func validateGetAccountRequest(req *pb.GetAccountRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateAccountId(req.GetId()); err != nil {
		violations = append(violations, fieldViolation("AccountId", err))
	}

	return
}
