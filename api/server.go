package api

import (
	"fmt"

	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/dubass83/simplebank/token"
	"github.com/dubass83/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewJwtMaker(config.TokenString)
	if err != nil {
		return nil, fmt.Errorf("can not create token maker: %w", err)
	}
	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()
	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)
	router.POST("/tokens/renew_access", server.renewAccessToken)

	authRouters := router.Group("/").Use(AuthMidleWare(server.tokenMaker))

	authRouters.GET("/users/:username", server.getUser)
	authRouters.POST("/accounts", server.createAccount)
	authRouters.GET("/accounts/:id", server.getAccount)
	authRouters.GET("/accounts", server.listAccount)
	authRouters.PUT("/accounts/:id", server.updateAccount)
	authRouters.DELETE("/accounts/:id", server.deleteAccount)
	authRouters.POST("/transfers", server.createTransfer)

	server.router = router
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func (srv *Server) Start(address string) error {
	return srv.router.Run(address)
}
