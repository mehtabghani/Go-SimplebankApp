package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/mehtabghani/simplebank/api"
	db "github.com/mehtabghani/simplebank/db/sqlc"
	"github.com/mehtabghani/simplebank/gapi"
	"github.com/mehtabghani/simplebank/pb"
	"github.com/mehtabghani/simplebank/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	config, err := util.LoadConfig(".")

	if err != nil {
		log.Fatal("cannot load config")
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
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
		log.Fatal("cannot create server")
	}

	// gprcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer()
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)

	if err != nil {
		log.Fatal("cannot create listener")
	}

	log.Printf("start gRPC server at %s", listener.Addr().String())

	err = grpcServer.Serve(listener)

	if err != nil {
		log.Fatal("cannot start gRPC server")
	}
}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)

	if err != nil {
		log.Fatal("cannot create server")
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot start server", err)
	}
}

// func runGatewayServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) {
func runGatewayServer(config util.Config, store db.Store) {

	// server, err := gapi.NewServer(config, store, taskDistributor)
	server, err := gapi.NewServer(config, store)

	if err != nil {
		log.Fatal("cannot create server")
	}

	// data in response is returning in camel case to turn it to your custom case enable this option
	// jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
	// 	MarshalOptions: protojson.MarshalOptions{
	// 		UseProtoNames: true,
	// 	},
	// 	UnmarshalOptions: protojson.UnmarshalOptions{
	// 		DiscardUnknown: true,
	// 	},
	// })

	// grpcMux := runtime.NewServeMux(jsonOption)
	grpcMux := runtime.NewServeMux()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal("cannot register handler server")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	// statikFS, err := fs.New()
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("cannot create statik fs")
	// }

	// swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	// mux.Handle("/swagger/", swaggerHandler)

	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot create listener")
	}

	log.Printf("start HTTP gateway server at %s", listener.Addr().String())
	// handler := gapi.HttpLogger(mux)
	// err = http.Serve(listener, handler)
	err = http.Serve(listener, mux)

	if err != nil {
		log.Fatal("cannot start HTTP gateway server")
	}
}
