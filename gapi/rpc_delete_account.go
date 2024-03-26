package gapi

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dubass83/simplebank/pb"
	"github.com/dubass83/simplebank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (srv *Server) DeleteAccount(ctx context.Context, req *pb.DeleteAccountRequest) (*pb.DeleteAccountResponse, error) {
	payload, err := srv.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if violations := validateDeleteAccountRequest(req); violations != nil {
		return nil, invalidArgumentError(violations)
	}

	Account, err := srv.store.GetAccount(ctx, req.GetId())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "account not found: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "cannot get Account: %s", err)
	}

	if payload.Username != Account.Owner {
		return nil, status.Errorf(
			codes.Unauthenticated,
			"user: %s not allouwd to Delete account ID: %d",
			payload.Username,
			req.GetId())
	}

	err = srv.store.DeleteAccount(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot Delete Account: %s", err)
	}

	rsp := &pb.DeleteAccountResponse{
		AccountWasDeleted: fmt.Sprint(req.GetId()),
	}
	return rsp, nil
}

func validateDeleteAccountRequest(req *pb.DeleteAccountRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateAccountId(req.GetId()); err != nil {
		violations = append(violations, fieldViolation("AccountId", err))
	}

	return
}
