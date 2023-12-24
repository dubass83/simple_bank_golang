package api

import (
	"database/sql"
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

type getAccountRequest struct {
	Id int64 `uri:"id" binding:"required,min=1"`
}

func (srv *Server) getAccount(ctx *gin.Context) {
	var req getAccountRequest
	err := ctx.ShouldBindUri(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := srv.store.GetAccount(ctx, req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, account)
}

type listAccountRequest struct {
	PageNumber int32 `form:"page_number" binding:"required,min=1"`
	PageSize   int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (srv *Server) listAccount(ctx *gin.Context) {
	var req listAccountRequest
	err := ctx.ShouldBind(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListAccountsParams{
		Limit:  req.PageSize,
		Offset: (req.PageNumber - 1) * req.PageSize,
	}

	accounts, err := srv.store.ListAccounts(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, accounts)
}

type updateAccountRequest struct {
	Balance int64 `json:"balance" binding:"required,min=0"`
}

type updateAccountId struct {
	Id int64 `uri:"id" binding:"required,min=1"`
}

func (srv *Server) updateAccount(ctx *gin.Context) {
	var reqPost updateAccountRequest
	err := ctx.ShouldBindJSON(&reqPost)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	var reqUri updateAccountId
	err = ctx.ShouldBindUri(&reqUri)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateAccountParams{
		ID:      reqUri.Id,
		Balance: reqPost.Balance,
	}

	account, err := srv.store.UpdateAccount(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, account)
}

type deleteAccountId struct {
	Id int64 `uri:"id" binding:"required,min=1"`
}

func (srv *Server) deleteAccount(ctx *gin.Context) {
	var reqUri updateAccountId
	err := ctx.ShouldBindUri(&reqUri)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err = srv.store.DeleteAccount(ctx, reqUri.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"accout was deleted": reqUri.Id})
}
