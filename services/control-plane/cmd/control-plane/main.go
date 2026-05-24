package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"eduadcrm/services/control-plane/internal/app"
	"eduadcrm/services/control-plane/internal/config"
	"eduadcrm/services/control-plane/internal/store"
)

func main() {
	configPath := getenv("EDUADCRM_CONFIG_FILE", "config/defaults.yaml")
	localConfigPath := getenv("EDUADCRM_LOCAL_CONFIG_FILE", "config/local.yaml")

	loader := config.NewFileLoader(configPath, localConfigPath, os.Getenv)
	manager, err := config.NewManager(loader)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	server, cleanup, err := newServer(manager)
	if err != nil {
		log.Fatalf("create server: %v", err)
	}
	defer cleanup()

	addr := getenv("CONTROL_PLANE_ADDR", ":8080")
	log.Printf("control-plane listening on %s", addr)
	if err := http.ListenAndServe(addr, server.Router()); err != nil {
		log.Fatal(err)
	}
}

func newServer(manager *config.Manager) (*app.Server, func(), error) {
	var repository store.Repository
	var cleanup func()
	if getenv("CONTROL_PLANE_STORE", "memory") != "postgres" {
		repository = store.NewMemoryStore()
		cleanup = func() {}
	} else {
		databaseURL := getenv("DATABASE_URL", "")
		if databaseURL == "" {
			return nil, nil, os.ErrInvalid
		}
		postgresStore, err := store.NewPostgresStore(context.Background(), databaseURL)
		if err != nil {
			return nil, nil, err
		}
		repository = postgresStore
		cleanup = postgresStore.Close
	}
	if redisAddr := getenv("REDIS_ADDR", ""); redisAddr != "" {
		repository = store.NewBlacklistRepository(repository, store.NewRedisBlacklist(redisAddr))
	}
	return app.NewServerWithStore(manager, repository), cleanup, nil
}

func getenv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
