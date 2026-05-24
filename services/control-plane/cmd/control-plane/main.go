package main

import (
	"log"
	"net/http"
	"os"

	"eduadcrm/services/control-plane/internal/app"
	"eduadcrm/services/control-plane/internal/config"
)

func main() {
	configPath := getenv("EDUADCRM_CONFIG_FILE", "config/defaults.yaml")
	localConfigPath := getenv("EDUADCRM_LOCAL_CONFIG_FILE", "config/local.yaml")

	loader := config.NewFileLoader(configPath, localConfigPath, os.Getenv)
	manager, err := config.NewManager(loader)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	server := app.NewServer(manager)
	addr := getenv("CONTROL_PLANE_ADDR", ":8080")
	log.Printf("control-plane listening on %s", addr)
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
