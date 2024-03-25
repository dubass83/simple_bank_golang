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

func (srv *Server) ListAccounts(ctx context.Context, req *pb.ListAccountsRequest) (*pb.ListAccountsResponse, error) {
	payload, err := srv.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if violation := validateListAccountRequest(req); violation != nil {
		return nil, invalidArgumentError(violation)
	}

	arg := db.ListAccountsParams{
		Owner:  payload.Username,
		Limit:  req.GetPageSize(),
		Offset: (req.GetPageNumber() - 1) * req.GetPageSize(),
	}

	accounts, err := srv.store.ListAccounts(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot get List of Accounts: %s", err)
	}

	rsp := &pb.ListAccountsResponse{
		Accounts: convertAccounts(accounts),
	}

	return rsp, nil
}

func validateListAccountRequest(req *pb.ListAccountsRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidatePageNumber(req.GetPageNumber()); err != nil {
		violations = append(violations, fieldViolation("AccountListPageNumber", err))
	}
	if err := val.ValidatePageSize(req.GetPageSize()); err != nil {
		violations = append(violations, fieldViolation("AccountListPageSize", err))
	}
	return
}
