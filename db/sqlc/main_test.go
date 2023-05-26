package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/mehtabghani/simplebank/util"
	// "github.com/techschool/simplebank/util"
)

const (
	dbDrive  = "postgres"
	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
)

var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")

	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}

// func TestMain(m *testing.M) {
// 	config, err := util.LoadConfig("../..")
// 	if err != nil {
// 		log.Fatal("cannot load config:", err)
// 	}

// 	testDB, err = sql.Open(config.DBDriver, config.DBSource)
// 	if err != nil {
// 		log.Fatal("cannot connect to db:", err)
// 	}

// 	testQueries = New(testDB)

// 	os.Exit(m.Run())
// }
