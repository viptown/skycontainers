package main

import (
	"log"
	"net/http"
	"os"

	"skycontainers/internal/auth"
	"skycontainers/internal/http/router"
	"skycontainers/internal/repo"
	"skycontainers/internal/view"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using defaults")
	}

	repo.InitDB()
	auth.InitAuth()
	view.InitTemplates()

	r := router.NewRouter()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}

