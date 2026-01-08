package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=public", user, pass, host, port, dbname)
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(ctx)

	log.Printf("Connected to database: %s\n", dbname)

	sqlFile, err := os.ReadFile("database.sql")
	if err != nil {
		log.Fatalf("Error reading database.sql: %v\n", err)
	}

	// Split by semicolon might be risky but simple for this file
	// A better way is to execute the whole block if possible
	queries := string(sqlFile)

	_, err = conn.Exec(ctx, queries)
	if err != nil {
		log.Fatalf("Error executing schema: %v\n", err)
	}

	log.Println("Schema created successfully")

	// Create admin user
	// Check if users table exists first (it should now)
	adminHash := "$2a$10$Zk9pPrpuG9eoHJjGSHISue49sjId52x/GT/8GtAUnOBi62G1kqv7C" // 12345

	// Insert initial supplier and admin
	initSQL := `
	INSERT INTO "suppliers" (id, name, tel, email, is_active, created_at, updated_at)
	VALUES (1, 'Sky Containers', '02-1234-5678', 'contact@sky.com', true, now(), now())
	ON CONFLICT (id) DO NOTHING;

	INSERT INTO "users" (id, supplier_id, uid, password_hash, name, email, duty, phone, role, status, last_login_at, created_at, updated_at)
	VALUES (1, 1, 'admin', '` + adminHash + `', 'Administrator', 'admin@sky.com', 'Senior Manager', '010-1111-2222', 'INTERNAL_SUPER_ADMIN', 'active', now(), now(), now())
	ON CONFLICT (id) DO NOTHING;

	SELECT setval(pg_get_serial_sequence('suppliers', 'id'), (SELECT COALESCE(MAX(id), 1) FROM suppliers));
	SELECT setval(pg_get_serial_sequence('users', 'id'), (SELECT COALESCE(MAX(id), 1) FROM users));
	`
	_, err = conn.Exec(ctx, initSQL)
	if err != nil {
		log.Fatalf("Error inserting initial data: %v\n", err)
	}

	log.Println("Initial data inserted successfully")
}
