package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"

	"gitlab.com/infra.run/public/b3scale/pkg/events"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// Get configuration from environment with
// a default fallback.
func getopt(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func main() {
	fmt.Println("b3scale event monitor		v.0.1.0")
	redisURL := getopt(
		"BBB_REDIS_URL", "redis://localhost:6379/1")
	dbConnStr := getopt(
		"B3SCALE_DB_URL",
		"postgres://postgres:postgres@localhost:5432/b3scale")
	// Initialize postgres connection
	dbConn, err := store.Connect(dbConnStr)
	if err != nil {
		log.Fatal(err)
	}

	// Make redis client
	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal(err)
	}

	rdb := redis.NewClient(redisOpts)
	monitor := events.NewMonitor(rdb)
	handler := NewEventHandler(dbConn)
	channel := monitor.Subscribe()
	for ev := range channel {
		err := handler.Dispatch(ev)
		if err != nil {
			log.Println("event handler error:", err)
		}
	}

}
