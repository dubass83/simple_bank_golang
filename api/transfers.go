package api

import (
	"errors"
	"fmt"
	"net/http"

	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/dubass83/simplebank/token"
	"github.com/gin-gonic/gin"
)

type createTransferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Carrency      string `json:"carrency" binding:"required,currency"`
}

func (srv *Server) createTransfer(ctx *gin.Context) {
	var req createTransferRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	fromAccount, ok := srv.validAccount(ctx, req.FromAccountID, req.Carrency)
	if !ok {
		return
	}
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Username != fromAccount.Owner {
		err := errors.New("fromAccount does not belong to authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	_, ok = srv.validAccount(ctx, req.ToAccountID, req.Carrency)
	if !ok {
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Ammount:       req.Amount,
	}

	transfer, err := srv.store.TransferTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, transfer)
}

func (srv *Server) validAccount(ctx *gin.Context, accountID int64, currency string) (db.Account, bool) {
	account, err := srv.store.GetAccount(ctx, accountID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return account, false
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return account, false
	}

	if account.Carrency != currency {
		err := fmt.Errorf("account [%d] currency: %s does not match - %s", account.ID, account.Carrency, currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return account, false
	}

	return account, true
}
