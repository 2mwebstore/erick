// Package database opens and verifies the MySQL connection.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kongchansila/portfolio/backend/internal/config"

	_ "github.com/go-sql-driver/mysql"
)

// Connect opens a pool and verifies it with a ping.
//
// The pool is bounded: an unbounded one lets a traffic spike open more
// connections than MySQL's max_connections allows, which turns a slow request
// into an outage.
func Connect(ctx context.Context, cfg config.DBConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("database: open: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnLifetime)

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		// The error is wrapped without the DSN so credentials never reach a log.
		return nil, fmt.Errorf("database: ping %s:%s/%s: %w", cfg.Host, cfg.Port, cfg.Name, err)
	}

	return db, nil
}
