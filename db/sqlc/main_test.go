package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/dubass83/simplebank/util"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testStore Store

func TestMain(m *testing.M) {
	conf, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("can not read config", err)
	}

	connPool, err := pgxpool.New(context.Background(), conf.DBSource)
	if err != nil {
		log.Fatal("can not connect to db", err)
	}
	testStore = NewStore(connPool)

	os.Exit(m.Run())
}
