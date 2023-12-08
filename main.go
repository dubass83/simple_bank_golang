package main

import (
	"database/sql"
	"log"

	"github.com/dubass83/simplebank/api"
	db "github.com/dubass83/simplebank/db/sqlc"
	_ "github.com/lib/pq"
)

const (
	dbDriver      = "postgres"
	dbSource      = "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable"
	addressString = "0.0.0.0:8080"
)

func main() {

	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("can not connect to db", err)
	}
	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(addressString)
	if err != nil {
		log.Fatal("can not start server", err)
	}
}
