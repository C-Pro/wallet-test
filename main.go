package main

import (
	"log"
	"net/http"
	"os"

	"github.com/c-pro/wallet-test/models"
)

func main() {
	url := os.Getenv("POSTGRESCONNSTR")
	if url == "" {
		url = "postgres://wallet:wallet@localhost:5432/wallet?sslmode=disable"
	}
	db, err := models.OpenDB(url)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.Handle("/", makeHandlers(db))

	http.Handle("/", mux)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
