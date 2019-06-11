package models

import (
	"context"
	"database/sql"
	"time"

	// blind import postgres driver
	_ "github.com/lib/pq"
)

const queryTimeout = time.Second * 5

// OpenDB connects to postgresql database
func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(0)
	db.SetMaxOpenConns(20)

	return db, nil
}
