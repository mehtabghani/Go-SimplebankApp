package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/mehtabghani/simplebank/api"
	db "github.com/mehtabghani/simplebank/db/sqlc"
	"github.com/mehtabghani/simplebank/util"
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
	server, err := api.NewServer(store)

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot start server", err)
	}
}
