package models

import (
	"context"
	"database/sql"
	"time"

	// blind import postgres driver
	_ "github.com/lib/pq"
)

const queryTimeout = time.Second * 5
const numRetries = 5
const retryDelay = time.Second * 2

// OpenDB connects to postgresql database
func OpenDB(dsn string) (*sql.DB, error) {
	var (
		db  *sql.DB
		err error
	)

	for i := 0; i < numRetries; i++ {
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			time.Sleep(retryDelay)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
		err = db.PingContext(ctx)
		cancel()
		if err != nil {
			time.Sleep(retryDelay)
			continue
		}
		break
	}

	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(0)
	db.SetMaxOpenConns(20)

	return db, nil
}
