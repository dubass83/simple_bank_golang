package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/dubass83/simplebank/api"
	db "github.com/dubass83/simplebank/db/sqlc"
	_ "github.com/dubass83/simplebank/docs/statik"
	"github.com/dubass83/simplebank/gapi"
	"github.com/dubass83/simplebank/pb"
	"github.com/dubass83/simplebank/util"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/rakyll/statik/fs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
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
	go runGateWayServer(conf, store)
	runGRPCServer(conf, store)
}

// runGateWayServer run Gateway server
func runGateWayServer(conf util.Config, store db.Store) {
	server, err := gapi.NewServer(conf, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	jsonOptions := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	grpcMux := runtime.NewServeMux(jsonOptions)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal("cannot register handler server:", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal("cannot create statik filesytem:", err)
	}

	swagerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	mux.Handle("/swagger/", swagerHandler)

	listener, err := net.Listen("tcp", conf.HTTPAddressString)
	if err != nil {
		log.Fatal("cannot create listener:", err)
	}
	log.Printf("start Gateway server on port %s", listener.Addr().String())
	err = http.Serve(listener, mux)
	if err != nil {
		log.Fatal("cannot start Gateway server:", err)
	}
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
