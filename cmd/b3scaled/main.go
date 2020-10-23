package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"

	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
	"gitlab.com/infra.run/public/b3scale/pkg/iface/http"
	"gitlab.com/infra.run/public/b3scale/pkg/middlewares/requests"
	"gitlab.com/infra.run/public/b3scale/pkg/middlewares/routing"
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
	banner() // Most important.
	log.Println("Booting b3scaled...")

	// Config
	listenHTTP := getopt(
		"B3SCALE_LISTEN_HTTP", "127.0.0.1:42353") // B3S
	dbConnStr := getopt(
		"B3SCALE_DB_URL",
		"postgres://postgres:postgres@localhost:5432/b3scale")

	// Initialize postgres connection
	dbConn, err := pgxpool.Connect(context.Background(), dbConnStr)
	if err != nil {
		log.Println("Error while connecting to database:", err)
		return
	}

	// Initialize cluster
	state := store.NewClusterState(dbConn)
	ctrl := cluster.NewController(state, dbConn)

	// Start router
	router := cluster.NewRouter(ctrl)
	router.Use(routing.Lookup(ctrl))
	// router.Use(routing.SortLoad)
	// router.Use(routing.SessionTracking...)

	// Start cluster request handler, and apply middlewares.
	// The middlewares are executes in reverse order.
	gateway := cluster.NewGateway()
	// gateway.Use(frontendFilter)
	gateway.Use(requests.NewDispatchMerge())
	gateway.Use(router.Middleware())
	go gateway.Start()

	// Start HTTP interface
	ifaceHTTP := http.NewInterface(
		listenHTTP,
		gateway)

	ifaceHTTP.Start()
}
