package main

import (
	"fmt"
	"log"
	"net/http"
	"recruitment-system/api"
	"recruitment-system/config"
	"recruitment-system/db"
)

func main() {
	// Load configuration from .env file
	config.LoadConfig()

	// Initialize database connection and run migrations
	db.Init()

	// Set up the router
	router := api.NewRouter()

	// Start the server
	port := config.AppConfig.ServerPort
	log.Printf("Server starting on port %s...", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), router)
	if err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
