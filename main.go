package main

import (
	"database/sql"
	"log"
	"net"

	"github.com/dubass83/simplebank/api"
	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/dubass83/simplebank/gapi"
	"github.com/dubass83/simplebank/pb"
	"github.com/dubass83/simplebank/util"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	conf, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("can not read config:", err)
	}

	conn, err := sql.Open(conf.DBDriver, conf.DBSource)
	if err != nil {
		log.Fatal("can not connect to db:", err)
	}
	store := db.NewStore(conn)
	runGRPCServer(conf, store)
}

// runGRPCServer run gRPC server
func runGRPCServer(conf util.Config, store db.Store) {
	server, err := gapi.NewServer(conf, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", conf.GRPCAddressString)
	if err != nil {
		log.Fatal("cannot create listener:", err)
	}
	log.Printf("start gRPC server on port %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start gRPC server")
	}
}

// runGinServer run http server with Gin framework
func runGinServer(conf util.Config, store db.Store) {
	server, err := api.NewServer(conf, store)
	if err != nil {
		log.Fatal("can not create server:", err)
	}
	err = server.Start(conf.HTTPAddressString)
	if err != nil {
		log.Fatal("can not start server:", err)
	}
}
