package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/mehtabghani/simplebank/api"
	db "github.com/mehtabghani/simplebank/db/sqlc"
	"github.com/mehtabghani/simplebank/gapi"
	"github.com/mehtabghani/simplebank/pb"
	"github.com/mehtabghani/simplebank/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	config, err := util.LoadConfig(".")

	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}

	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}

	store := db.NewStore(conn)
	// runGinServer(config, store)

	go runGatewayServer(config, store)

	runGrpcServer(config, store)
}

func runGrpcServer(config util.Config, store db.Store) {
	// server, err := gapi.NewServer(config, store, taskDistributor)
	server, err := gapi.NewServer(config, store)

	if err != nil {
		log.Fatal().Msg("cannot create server")
	}

	gprcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger) // GRPC Interceptor
	grpcServer := grpc.NewServer(gprcLogger)
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)

	if err != nil {
		log.Fatal().Msg("cannot create listener")
	}

	log.Info().Msgf("start gRPC server at %s", listener.Addr().String())

	err = grpcServer.Serve(listener)

	if err != nil {
		log.Fatal().Msg("cannot start gRPC server")
	}
}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)

	if err != nil {
		log.Fatal().Msg("cannot create server")
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}

// func runGatewayServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) {
func runGatewayServer(config util.Config, store db.Store) {

	// server, err := gapi.NewServer(config, store, taskDistributor)
	server, err := gapi.NewServer(config, store)

	if err != nil {
		log.Fatal().Msg("cannot create server")
	}

	// data in REST apyresponse is returning in camel case to turn it to your custom case enable this option
	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	grpcMux := runtime.NewServeMux(jsonOption)
	// grpcMux := runtime.NewServeMux()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal().Msg("cannot register handler server")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	// TODO: need to revisit lecture: Sec 4, 46
	// statikFS, err := fs.New()
	// if err != nil {
	// 	log.Fatal().Msg().Err(err).Msg("cannot create statik fs")
	// }

	// swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	// mux.Handle("/swagger/", swaggerHandler)

	// This will allow to run swagger locally
	// at http://localhost:8080/swagger/
	fs := http.FileServer(http.Dir("./doc/swagger"))
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", fs))

	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Msg("cannot create listener")
	}

	log.Info().Msgf("start HTTP gateway server at %s", listener.Addr().String())
	handler := gapi.HttpLogger(mux) // this will return new HTTP handler with a logger middleware
	err = http.Serve(listener, handler)
	// err = http.Serve(listener, mux)

	if err != nil {
		log.Fatal().Msg("cannot start HTTP gateway server")
	}
}
