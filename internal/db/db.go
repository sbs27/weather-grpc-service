package db

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ConnectDB initializes a connection pool to PostgreSQL
func ConnectDB() *pgxpool.Pool {
	// 1. Look for environment variable first (Docker style)
	// 2. Fall back to localhost (Local development style)
	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		connStr = "postgres://user:password@localhost:5432/weather_db"
	}

	// Create the Pool
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatalf("Unable to parse database URL: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	log.Println(" Successfully connected to PostgreSQL!")
	return pool
}
