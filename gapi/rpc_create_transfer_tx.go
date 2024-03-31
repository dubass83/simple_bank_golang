package gapi

import (
	"context"
	"fmt"

	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/dubass83/simplebank/pb"
	"github.com/dubass83/simplebank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (srv *Server) CreateTransfer(ctx context.Context, req *pb.CreateTransferTxRequest) (*pb.CreateTransferTxResponse, error) {
	payload, err := srv.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if violation := validateCreateTransferRequest(req); violation != nil {
		return nil, invalidArgumentError(violation)
	}

	fromAccount, err := srv.store.GetAccount(ctx, req.GetFromAccountId())
	if err != nil {
		return nil, cannotGetAccountError(err)
	}

	toAccount, err := srv.store.GetAccount(ctx, req.GetToAccountId())
	if err != nil {
		return nil, cannotGetAccountError(err)
	}

	if err = val.ValidateTxCarrency(fromAccount, toAccount); err != nil {
		violations := []*errdetails.BadRequest_FieldViolation{}
		violations = append(violations, fieldViolation("TxCarrency", err))
		return nil, invalidArgumentError(violations)
	}

	account, err := srv.store.GetAccount(ctx, req.GetFromAccountId())
	if err != nil {
		return nil, fmt.Errorf("can not get account from id: %s", err)
	}
	if account.Owner != payload.Username {
		return nil, fmt.Errorf("user allowed to transfer money only from his account")
	}

	arg := db.TransferTxParams{
		FromAccountID: req.GetFromAccountId(),
		ToAccountID:   req.GetToAccountId(),
		Ammount:       req.Amount,
	}

	txResp, err := srv.store.TransferTx(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create Transfer: %s", err)
	}

	rsp := convertTransferTx(txResp)

	return rsp, nil
}

func validateCreateTransferRequest(req *pb.CreateTransferTxRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateAccountId(req.GetFromAccountId()); err != nil {
		violations = append(violations, fieldViolation("FromAccountId", err))
	}
	if err := val.ValidateAccountId(req.GetToAccountId()); err != nil {
		violations = append(violations, fieldViolation("ToAccountId", err))
	}
	if err := val.ValidateMoneyAmmount(req.GetAmount()); err != nil {
		violations = append(violations, fieldViolation("MoneyAmount", err))
	}
	return
}
