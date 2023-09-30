package store

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"

	// TODO: clean up this imports mess.
	msqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/server"
)

const (
	testDatabaseName    = "tir"
	testDatabaseAddress = "127.0.0.1"
	testDatabasePort    = 3306
)

type storeBuilder func(tb testing.TB) Interface

func BuildMemory(tb testing.TB) Interface {
	return UseMemory()
}

func BuildFile(tb testing.TB) Interface {
	file, err := os.CreateTemp(tb.TempDir(), "*.json")
	if err != nil {
		tb.Fatalf("error creating temp file: %v", err)
	}
	store, err := UseFile(file.Name())
	if err != nil {
		tb.Fatalf("error using file as store: %v", err)
	}
	return store
}

func BuildMySQL(tb testing.TB) Interface {
	engine := msqle.NewDefault(memory.NewDBProvider(memory.NewDatabase(testDatabaseName)))
	config := server.Config{
		Protocol: "tcp",
		Address:  fmt.Sprintf("%s:%d", testDatabaseAddress, testDatabasePort),
	}
	s, err := server.NewDefaultServer(config, engine)
	if err != nil {
		tb.Fatalf("Failed creating MySQL server: %v", err)
	}
	go s.Start()
	tb.Cleanup(func() {
		if err := s.Close(); err != nil {
			tb.Logf("Error closing server: %v", err)
		}
	})

	db, err := sql.Open("mysql", fmt.Sprintf("tcp(%s:%d)/%v?parseTime=true", testDatabaseAddress, testDatabasePort, testDatabaseName))
	if err != nil {
		tb.Fatalf("Failed opening DB: %v", err)
	}

	store, err := UseSQL(db)
	if err != nil {
		tb.Fatalf("error initializing store: %v", err)
	}
	return store
}
