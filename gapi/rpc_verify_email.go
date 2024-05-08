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

func (srv *Server) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	if violation := validateVerifyEmailRequest(req); violation != nil {
		return nil, invalidArgumentError(violation)
	}

	arg := db.VerifyEmailTxParams{
		ID:         req.Id,
		SecretCode: req.SecretCode,
	}
	resultTx, err := srv.store.VerifyEmailTx(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot validate user email: %s", err)
	}

	rsp := &pb.VerifyEmailResponse{IsVerified: resultTx.User.IsEmailVerified}

	return rsp, nil
}

func validateVerifyEmailRequest(req *pb.VerifyEmailRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateVerifyEmailID(req.GetId()); err != nil {
		violations = append(violations, fieldViolation("verify_email_id", err))
	}
	if err := val.ValidateVerifyEmailSecretCode(req.GetSecretCode()); err != nil {
		violations = append(violations, fieldViolation("verify_email_secret_code", err))
	}
	return
}
