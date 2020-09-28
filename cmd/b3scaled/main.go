package main

import (
	"log"
	"os"

	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
	"gitlab.com/infra.run/public/b3scale/pkg/config"
	"gitlab.com/infra.run/public/b3scale/pkg/iface/http"
	"gitlab.com/infra.run/public/b3scale/pkg/middlewares/requests"
	// "gitlab.com/infra.run/public/b3scale/pkg/middlewares/routing"
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
	log.Println("Starting b3scaled.")

	// Config
	frontendsConfigFilename := getopt(
		"B3SCALE_FRONTENDS", "etc/b3scale/frontends.conf")
	backendsConfigFilename := getopt(
		"B3SCALE_BACKENDS", "etc/b3scale/backends.conf")
	listenHTTP := getopt(
		"B3SCALE_LISTEN_HTTP", "127.0.0.1:42353") // B3S

	log.Println("Using frontends from:", frontendsConfigFilename)
	log.Println("Using backends from:", backendsConfigFilename)

	// Initialize configuration
	backendsConfig := config.NewBackendsFileConfig(
		backendsConfigFilename)
	frontendsConfig := config.NewFrontendsFileConfig(
		frontendsConfigFilename)

	// Initialize cluster
	state := cluster.NewState()
	go state.Start()

	// Start cluster controller
	controller := cluster.NewController(
		state,
		backendsConfig,
		frontendsConfig)
	go controller.Start()

	// Start control signal handler
	ctl := NewSigCtl(controller)
	go ctl.Start()

	// Start router
	router := cluster.NewRouter(state)
	// router.Use(routing.SortLoad)
	// router.Use(routing.StickyBackends...)

	// Start cluster request handler, and apply middlewares.
	// The middlewares are executes in reverse order.
	gateway := cluster.NewGateway(state)
	// gateway.Use(frontendFilter)
	gateway.Use(requests.NewDispatchMerge())
	// gateway.Use(cache)
	gateway.Use(router.Middleware())
	go gateway.Start()

	// Start HTTP interface
	ifaceHTTP := http.NewInterface(
		listenHTTP,
		controller,
		gateway)

	ifaceHTTP.Start()
}
