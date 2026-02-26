package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ConnectDB initializes a connection pool to PostgreSQL
func ConnectDB() *pgxpool.Pool {
	// The Connection String (URL)
	connStr := "postgres://user:password@localhost:5432/weather_db"

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
