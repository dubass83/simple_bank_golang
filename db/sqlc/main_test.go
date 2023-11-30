package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable"
)

var TestQueries *Queries
var TestDb *sql.DB

func TestMain(m *testing.M) {
	var err error

	TestDb, err = sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("can not connect to db", err)
	}
	TestQueries = New(TestDb)

	os.Exit(m.Run())
}
