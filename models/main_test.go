package models

import (
	"database/sql"

	"os"
	"testing"
)

var db *sql.DB

func TestMain(m *testing.M) {
	url := os.Getenv("POSTGRESCONNSTR")
	if url == "" {
		url = "postgres://wallet:wallet@localhost:5432/wallet?sslmode=disable"
	}
	var err error
	db, err = OpenDB(url)

	if err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
