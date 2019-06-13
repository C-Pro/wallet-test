package main

import (
	"database/sql"
	"net/http/httptest"

	"os"
	"testing"

	"github.com/c-pro/wallet-test/models"
)

var (
	db  *sql.DB
	srv *httptest.Server
)

func TestMain(m *testing.M) {
	url := os.Getenv("POSTGRESCONNSTR")
	if url == "" {
		url = "postgres://wallet:wallet@localhost:5432/wallet?sslmode=disable"
	}
	var err error
	db, err = models.OpenDB(url)
	if err != nil {
		panic(url)
	}
	srv = httptest.NewServer(makeHandlers(db))
	os.Exit(m.Run())
}
