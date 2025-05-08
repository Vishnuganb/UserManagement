package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq" // It needs to use conjunction with a database driver
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:secret@localhost:5433/userManagement?sslmode=disable"
)

var testQueries *Queries

// The manager for running the entire test suite
func TestMain(m *testing.M) {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to the database:", err)
	}
	// Make Queries object so tests can use it
	testQueries = New(conn)

	// Run all the test cases (like TestCreateUser etc.)
	exitCode := m.Run()

	// Exit the program with correct test result code
	os.Exit(exitCode)
}
