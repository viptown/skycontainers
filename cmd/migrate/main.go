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
	_ = godotenv.Load()

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

	var regclass *string
	if err := conn.QueryRow(ctx, `SELECT to_regclass('public.container_types')`).Scan(&regclass); err != nil {
		log.Fatalf("Error checking schema state: %v\n", err)
	}

	if regclass == nil {
		sqlFile, err := os.ReadFile("database.sql")
		if err != nil {
			log.Fatalf("Error reading database.sql: %v\n", err)
		}

		if _, err := conn.Exec(ctx, string(sqlFile)); err != nil {
			log.Fatalf("Error executing schema: %v\n", err)
		}
		log.Println("Schema created successfully")
	} else {
		log.Println("Schema already exists, skipping schema creation")
	}

	adminHash := "$2a$10$Zk9pPrpuG9eoHJjGSHISue49sjId52x/GT/8GtAUnOBi62G1kqv7C" // 12345

	initSQL := `
        INSERT INTO "suppliers" (id, name, tel, email, is_active, created_at, updated_at)
        VALUES (1, 'Sky Containers', '02-1234-5678', 'contact@sky.com', true, now(), now())
        ON CONFLICT (id) DO NOTHING;

        INSERT INTO "users" (id, supplier_id, uid, password_hash, name, email, duty, phone, role, status, last_login_at, created_at, updated_at)
        VALUES (1, 1, 'admin', '` + adminHash + `', 'Administrator', 'admin@sky.com', 'Senior Manager', '010-1111-2222', 'INTERNAL_SUPER_ADMIN', 'active', now(), now(), now())
        ON CONFLICT (id) DO NOTHING;
        `
	if _, err := conn.Exec(ctx, initSQL); err != nil {
		log.Fatalf("Error inserting initial data: %v\n", err)
	}
	log.Println("Initial data inserted successfully")
}

