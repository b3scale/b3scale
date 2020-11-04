package main

import (
	"log"
	"os"

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
	dbConn, err := store.Connect(dbConnStr)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize cluster
	ctrl := cluster.NewController(dbConn)

	// Start router
	router := cluster.NewRouter(ctrl)
	router.Use(routing.Lookup(ctrl))
	// router.Use(routing.SortLoad)
	// router.Use(routing.SessionTracking...)

	// Start cluster request handler, and apply middlewares.
	// The middlewares are executes in reverse order.
	gateway := cluster.NewGateway(ctrl)
	// gateway.Use(frontendFilter)
	gateway.Use(requests.NewDispatchMerge())
	gateway.Use(router.Middleware())

	// Start cluster controller
	go ctrl.Start()

	// Start HTTP interface
	ifaceHTTP := http.NewInterface(
		listenHTTP,
		ctrl,
		gateway)

	ifaceHTTP.Start()
}
