package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type renewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type renewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_exp_at"`
}

func (srv *Server) renewAccessToken(ctx *gin.Context) {
	var req renewAccessTokenRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	payload, err := srv.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	session, err := srv.store.GetSession(ctx, uuid.UUID(payload.ID))
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if session.IsBloked {
		err := fmt.Errorf("session is bloked")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	if session.Username != payload.Username {
		err := fmt.Errorf("user in session does not much to user in refresh Payload")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	if session.RefreshToken != req.RefreshToken {
		err := fmt.Errorf("refresh token does not much")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	if time.Now().After(session.ExpiredAt) {
		err := fmt.Errorf("expired session")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	token, accessPayload, err := srv.tokenMaker.CreateToken(session.Username, srv.config.TokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := renewAccessTokenResponse{
		AccessToken:          token,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}

	ctx.JSON(http.StatusOK, rsp)
}
