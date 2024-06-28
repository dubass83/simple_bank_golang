package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dubass83/simplebank/api"
	db "github.com/dubass83/simplebank/db/sqlc"
	_ "github.com/dubass83/simplebank/docs/statik"
	"github.com/dubass83/simplebank/gapi"
	"github.com/dubass83/simplebank/mail"
	"github.com/dubass83/simplebank/pb"
	"github.com/dubass83/simplebank/util"
	"github.com/dubass83/simplebank/worker"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rakyll/statik/fs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

var interruptSignas = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

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

	ctx, stop := signal.NotifyContext(context.Background(), interruptSignas...)
	defer stop()

	connPool, err := pgxpool.New(ctx, conf.DBSource)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("method", "main").
			Msg("cannot validate db connaction")
	}
	store := db.NewStore(connPool)

	redisOpts := asynq.RedisClientOpt{
		Addr: conf.RedisAddress,
	}

	RedisTaskDestributor := worker.NewRedisTaskDistributor(redisOpts)

	runDbMigration(conf.MigrationURL, conf.DBSource)

	waitGroup, ctx := errgroup.WithContext(ctx)

	runTaskProcessor(ctx, waitGroup, conf, redisOpts, store)
	runGateWayServer(ctx, waitGroup, conf, store, RedisTaskDestributor)
	runGRPCServer(ctx, waitGroup, conf, store, RedisTaskDestributor)

	err = waitGroup.Wait()
	if err != nil {
		log.Fatal().Err(err).Msg("error from wait group ")
	}
}

// runDbMigration run db migration from provided URL to the  db
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
func runGateWayServer(
	ctx context.Context,
	waitGroup *errgroup.Group,
	conf util.Config,
	store db.Store,
	taskDistributor worker.TaskDistributor,
) {
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

	httpServer := &http.Server{
		Handler: gapi.HttpLogger(mux),
		Addr:    conf.HTTPAddressString,
	}
	waitGroup.Go(func() error {
		log.Info().Msgf("start Gateway server at %s", httpServer.Addr)
		err = httpServer.ListenAndServe()
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			log.Error().
				Err(err).
				Str("method", "main").
				Msg("cannot start Gateway server")
			return err
		}
		return nil
	})
	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("stoping http server")
		err := httpServer.Shutdown(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("error gracefull shutdown http server")
			return err
		}
		log.Info().Msg("http server is stoped")
		return nil
	})

}

// runTaskProcessor connect to redis and process task from the queue
func runTaskProcessor(
	ctx context.Context,
	waitGroup *errgroup.Group,
	conf util.Config,
	redisOpts asynq.RedisClientOpt,
	store db.Store,
) {
	sender := mail.NewMailtrapSender(
		conf.EmailSenderName,
		conf.EmailSenderEmailFrom,
		conf.MailtrapLogin,
		conf.MailtrapPass,
	)
	taskProcessor := worker.NewRedisTaskProcessor(redisOpts, store, sender)
	log.Info().Msg("start processing tasks from redis queue")
	err := taskProcessor.Start()
	if err != nil {
		log.Fatal().
			Err(err).
			Str("method", "main").
			Msg("cannot start taskProcessor")
	}
	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("gracefully stop redis task server")

		taskProcessor.Stop()
		log.Info().Msg("redis task server is stoped")
		return nil
	})
}

// runGRPCServer run gRPC server
func runGRPCServer(
	ctx context.Context,
	waitGroup *errgroup.Group,
	conf util.Config,
	store db.Store,
	taskDistributor worker.TaskDistributor,
) {
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
	waitGroup.Go(func() error {
		log.Info().Msgf("start gRPC server on port %s", listener.Addr().String())
		err = grpcServer.Serve(listener)
		if err != nil {
			if errors.Is(err, grpc.ErrServerStopped) {
				return nil
			}
			log.Error().
				Err(err).
				Str("method", "main").
				Msg("cannot start gRPC server")
			return err
		}
		return nil
	})
	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("graceful stop gRPC server")
		grpcServer.GracefulStop()
		log.Info().Msg("gRPC server is stoped")
		return nil
	})
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
