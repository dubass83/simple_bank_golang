package gapi

import (
	"context"

	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/dubass83/simplebank/pb"
	"github.com/dubass83/simplebank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (srv *Server) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {
	payload, err := srv.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if violation := validateCreateAccountRequest(req); violation != nil {
		return nil, invalidArgumentError(violation)
	}

	arg := db.CreateAccountParams{
		Owner:    payload.Username,
		Carrency: req.GetCurrency(),
		Balance:  0,
	}

	Account, err := srv.store.CreateAccount(ctx, arg)
	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolation {
			return nil, status.Errorf(
				codes.AlreadyExists, "Account for carrancy %s already exist: %s",
				req.GetCurrency(),
				err)
		}
		return nil, status.Errorf(codes.Internal, "cannot create Account: %s", err)
	}

	rsp := &pb.CreateAccountResponse{
		Account: convertAccount(Account),
	}

	return rsp, nil
}

func validateCreateAccountRequest(req *pb.CreateAccountRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateCurrency(req.GetCurrency()); err != nil {
		violations = append(violations, fieldViolation("Currency", err))
	}
	return
}
