package gapi

import (
	"context"
	"time"

	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/dubass83/simplebank/pb"
	"github.com/dubass83/simplebank/util"
	"github.com/dubass83/simplebank/val"
	"github.com/dubass83/simplebank/worker"
	"github.com/hibiken/asynq"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (srv *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	if violation := validateCreateUserRequest(req); violation != nil {
		return nil, invalidArgumentError(violation)
	}
	hash, err := util.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create hash from password: %s", err)
	}

	arg := db.CreateUserTxParams{
		CreateUserParams: db.CreateUserParams{
			Username:       req.GetUsername(),
			HashedPassword: hash,
			FullName:       req.GetFullName(),
			Email:          req.GetEmail(),
		},
		AfterFunc: func(user db.User) error {

			payload := &worker.PayloadSendVerifyEmail{
				Username: user.Username,
			}
			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(worker.QueueCritical),
			}
			return srv.taskDestributor.DestributeTaskSendVerifyEmail(ctx, payload, opts...)
		},
	}
	// log.Info().Msg(">> start creating user")
	// time.Sleep(time.Second * 10)

	userTx, err := srv.store.CreateUserTx(ctx, arg)
	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolation {
			return nil, status.Errorf(codes.AlreadyExists, "already exist: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "cannot create user: %s", err)
	}
	// log.Info().Msg(">> user created")
	rsp := &pb.CreateUserResponse{
		User: convertUser(userTx.User),
	}

	return rsp, nil
}

func validateCreateUserRequest(req *pb.CreateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}
	if err := val.ValidateFullname(req.GetFullName()); err != nil {
		violations = append(violations, fieldViolation("full_name", err))
	}
	if err := val.ValidateEmail(req.GetEmail()); err != nil {
		violations = append(violations, fieldViolation("email", err))
	}
	if err := val.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}
	return
}
