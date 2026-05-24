package main

import (
	"log"
	"net/http"
	"os"

	"eduadcrm/services/ad-integration/internal/app"
	"eduadcrm/services/ad-integration/internal/store"
)

func main() {
	repository := store.Repository(store.NewMemoryStore())
	if redisAddr := getenv("REDIS_ADDR", ""); redisAddr != "" {
		repository = store.NewStateRepository(repository, store.NewRedisStateStore(redisAddr))
	}
	var queue store.MessageQueue
	var cleanup func() = func() {}
	if databaseURL := getenv("DATABASE_URL", ""); databaseURL != "" {
		postgresQueue, err := store.NewPostgresQueue(databaseURL)
		if err != nil {
			log.Fatalf("create postgres queue: %v", err)
		}
		queue = postgresQueue
		cleanup = func() { _ = postgresQueue.Close() }
	}
	defer cleanup()
	server := app.NewServerWithStoreAndQueue(repository, queue)
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
