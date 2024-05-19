package gapi

import (
	"context"
	"database/sql"

	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/dubass83/simplebank/pb"
	"github.com/dubass83/simplebank/util"
	"github.com/dubass83/simplebank/val"
	"github.com/google/uuid"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (srv *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	if violations := validateLoginUserRequest(req); violations != nil {
		return nil, invalidArgumentError(violations)
	}
	user, err := srv.store.GetUser(ctx, req.GetUsername())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "cannot get user")
	}

	err = util.CheckPassword(req.GetPassword(), user.HashedPassword)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "password doesnot match")
	}

	accessToken, accessPayload, err := srv.tokenMaker.CreateToken(req.GetUsername(), srv.config.TokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed  create access  token")
	}

	refreshToken, refreshPayload, err := srv.tokenMaker.CreateToken(req.GetUsername(), srv.config.RefreshTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed create refresh token")
	}

	mtdt := srv.extractMetadata(ctx)

	session, err := srv.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           uuid.UUID(refreshPayload.ID),
		Username:     req.GetUsername(),
		RefreshToken: refreshToken,
		UserAgent:    mtdt.UserAgent,
		ClientIp:     mtdt.ClientIP,
		IsBloked:     false,
		ExpiredAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed create session")
	}

	rsp := &pb.LoginUserResponse{
		SessionId:         session.ID.String(),
		AccessToken:       accessToken,
		AccessTokenExpAt:  timestamppb.New(accessPayload.ExpiredAt),
		RefreshToken:      refreshToken,
		RefreshTokenExpAt: timestamppb.New(refreshPayload.ExpiredAt),
		User:              convertUser(user),
	}
	return rsp, nil
}

func validateLoginUserRequest(req *pb.LoginUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}
	if err := val.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}
	return
}
