package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"

	handler "go-bot-tele/api"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Map the handler function to the same path as Vercel
	http.HandleFunc("/api/handler", handler.Handler)

	log.Println("🚀 Local server running on http://localhost:8080")
	log.Println("Test Cron: curl \"http://localhost:8080/api/handler?mode=cron\"")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
