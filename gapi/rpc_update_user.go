package gapi

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/dubass83/simplebank/pb"
	"github.com/dubass83/simplebank/util"
	"github.com/dubass83/simplebank/val"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (srv *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	payload, err := srv.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if violation := validateUpdateUserRequest(req); violation != nil {
		return nil, invalidArgumentError(violation)
	}

	if payload.Username != req.GetUsername() {
		err := fmt.Errorf("user: %s is not authorized to make changes in another user: %s",
			payload.Username,
			req.GetUsername(),
		)
		return nil, unauthenticatedError(err)
	}

	arg := db.UpdateUserParams{
		Username: req.GetUsername(),
		FullName: pgtype.Text{
			String: req.GetFullName(),
			Valid:  req.FullName != nil,
		},
		Email: pgtype.Text{
			String: req.GetEmail(),
			Valid:  req.Email != nil,
		},
	}

	if req.Password != nil {
		hash, err := util.HashPassword(req.GetPassword())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "cannot Update hash from password: %s", err)
		}
		arg.HashedPassword = pgtype.Text{
			String: hash,
			Valid:  true,
		}
		arg.PasswordChangedAt = pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		}
	}

	user, err := srv.store.UpdateUser(ctx, arg)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "cannot find user in database: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "cannot Update user: %s", err)
	}

	rsp := &pb.UpdateUserResponse{
		User: convertUser(user),
	}

	return rsp, nil
}

func validateUpdateUserRequest(req *pb.UpdateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}
	if req.FullName != nil {
		if err := val.ValidateFullname(req.GetFullName()); err != nil {
			violations = append(violations, fieldViolation("full_name", err))
		}
	}
	if req.Email != nil {
		if err := val.ValidateEmail(req.GetEmail()); err != nil {
			violations = append(violations, fieldViolation("email", err))
		}
	}
	if req.Password != nil {
		if err := val.ValidatePassword(req.GetPassword()); err != nil {
			violations = append(violations, fieldViolation("password", err))
		}
	}
	return
}
