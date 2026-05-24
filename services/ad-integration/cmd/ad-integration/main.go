package main

import (
	"log"
	"net/http"
	"os"

	"eduadcrm/services/ad-integration/internal/app"
)

func main() {
	server := app.NewServer()
	addr := getenv("AD_INTEGRATION_ADDR", ":8081")
	log.Printf("ad-integration listening on %s", addr)
	if err := http.ListenAndServe(addr, server.Router()); err != nil {
		log.Fatal(err)
	}
}

func getenv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
