package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/dubass83/simplebank/util"
	_ "github.com/lib/pq"
)

var TestQueries *Queries
var TestDb *sql.DB

func TestMain(m *testing.M) {
	conf, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("can not read config", err)
	}

	TestDb, err = sql.Open(conf.DBDriver, conf.DBSource)
	if err != nil {
		log.Fatal("can not connect to db", err)
	}
	TestQueries = New(TestDb)

	os.Exit(m.Run())
}
