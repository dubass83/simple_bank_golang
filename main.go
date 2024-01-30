package main

import (
	"database/sql"
	"log"

	"github.com/dubass83/simplebank/api"
	db "github.com/dubass83/simplebank/db/sqlc"
	"github.com/dubass83/simplebank/util"
	_ "github.com/lib/pq"
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
	server, err := api.NewServer(conf, store)
	if err != nil {
		log.Fatal("can not create server:", err)
	}
	err = server.Start(conf.AddressString)
	if err != nil {
		log.Fatal("can not start server:", err)
	}
}
