package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"

	"github.com/dubass83/simplebank/api"
	db "github.com/dubass83/simplebank/db/sqlc"
	_ "github.com/dubass83/simplebank/docs/statik"
	"github.com/dubass83/simplebank/gapi"
	"github.com/dubass83/simplebank/pb"
	"github.com/dubass83/simplebank/util"
	"github.com/dubass83/simplebank/worker"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hibiken/asynq"
	_ "github.com/lib/pq"
	"github.com/rakyll/statik/fs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	// UNIX Time is faster and smaller than most timestamps
	// zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	conf, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().
			Err(err).
			Str("method", "main").
			Msg("can not read config")
	}

	if conf.Enviroment == "devel" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	conn, err := sql.Open(conf.DBDriver, conf.DBSource)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("method", "main").
			Msg("cannot validate db connaction")
	}
	store := db.NewStore(conn)

	redisOpts := asynq.RedisClientOpt{
		Addr: conf.RedisAddress,
	}

	RedisTaskDestributor := worker.NewRedisTaskDistributor(redisOpts)

	runDbMigration(conf.MigrationURL, conf.DBSource)

	go runTaskProcessor(redisOpts, store)
	go runGateWayServer(conf, store, RedisTaskDestributor)
	runGRPCServer(conf, store, RedisTaskDestributor)
}

// runDbMigration run db migration from provided URL to the db
func runDbMigration(migrationURL, dbSource string) {
	m, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("method", "main").
			Msg("cannot create migration instance")
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().
			Err(err).
			Str("method", "main").
			Msg("cannot run migration up")
	}
	log.Info().Msg("successfully run db migration")
}

// runGateWayServer run Gateway server
func runGateWayServer(conf util.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(conf, store, taskDistributor)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("method", "main").
			Msg("cannot create server")
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
		log.Fatal().
			Err(err).
			Str("method", "main").
			Msg("cannot register handler server")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal().
			Err(err).
			Str("method", "main").
			Msg("cannot create statik filesytem")
	}

	swagerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	mux.Handle("/swagger/", swagerHandler)

	listener, err := net.Listen("tcp", conf.HTTPAddressString)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("method", "main").
			Msg("cannot create listener")
	}
	log.Info().Msgf("start Gateway server on port %s", listener.Addr().String())
	handler := gapi.HttpLogger(mux)
	err = http.Serve(listener, handler)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("method", "main").
			Msg("cannot start Gateway server")
	}
}

// runTaskProcessor connect to redis and process task from the queue
func runTaskProcessor(redisOpts asynq.RedisClientOpt, store db.Store) {
	taskProcessor := worker.NewRedisTaskProcessor(redisOpts, store)
	log.Info().Msg("start processing tasks from redis queue")
	err := taskProcessor.Start()
	if err != nil {
		log.Fatal().
			Err(err).
			Str("method", "main").
			Msg("cannot start taskProcessor")
	}
}

// runGRPCServer run gRPC server
func runGRPCServer(conf util.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(conf, store, taskDistributor)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("method", "main").
			Msg("cannot create server")
	}

	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", conf.GRPCAddressString)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("method", "main").
			Msg("cannot create listener")
	}
	log.Info().Msgf("start gRPC server on port %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("method", "main").
			Msg("cannot start gRPC server")
	}
}

// runGinServer run http server with Gin framework
func runGinServer(conf util.Config, store db.Store) {
	server, err := api.NewServer(conf, store)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("method", "main").
			Msg("can not create server")
	}
	err = server.Start(conf.HTTPAddressString)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("method", "main").
			Msg("can not start server")
	}
}
