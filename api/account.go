package api

import (
	"net/http"

	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/gin-gonic/gin"
)

type createAccountRequest struct {
	Owner    string `json:"owner" binding:"required"`
	Carrency string `json:"carrency" binding:"required,oneof=UAH USD EUR"`
}

func (srv *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	arg := db.CreateAccountParams{
		Owner:    req.Owner,
		Carrency: req.Carrency,
		Balance:  0,
	}

	account, err := srv.store.CreateAccount(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, account)
}
