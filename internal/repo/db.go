package repo

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func InitDB() {
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=public", user, pass, host, port, dbname)

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatalf("Unable to parse connection string: %v\n", err)
	}

	// Set some pool settings
	config.MaxConns = 10
	config.MaxConnIdleTime = time.Minute * 5
	if config.ConnConfig.RuntimeParams == nil {
		config.ConnConfig.RuntimeParams = map[string]string{}
	}
	config.ConnConfig.RuntimeParams["search_path"] = "public"

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Could not ping database: %v\n", err)
	}

	DB = pool
	var currentDB, currentUser, currentSchema, serverAddr string
	err = pool.QueryRow(context.Background(), "SELECT current_database(), current_user, current_schema(), inet_server_addr()::text").Scan(&currentDB, &currentUser, &currentSchema, &serverAddr)
	if err != nil {
		log.Printf("Database connection established (context query failed: %v)", err)
	} else {
		log.Printf("Database connection established: db=%s user=%s schema=%s host=%s", currentDB, currentUser, currentSchema, serverAddr)
	}
}
