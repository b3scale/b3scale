package main

import (
	"log"
	"os"

	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
	"gitlab.com/infra.run/public/b3scale/pkg/config"
	"gitlab.com/infra.run/public/b3scale/pkg/iface/http"
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

	// Start cluster controller
	controller := cluster.NewController(
		backendsConfig, frontendsConfig)
	go controller.Start()

	// Start cluster request handler
	gateway := cluster.NewGateway(controller)
	go gateway.Start()

	// Start control signal handler
	ctl := NewSigCtl(controller)
	go ctl.Start()

	// Start HTTP interface
	ifaceHTTP := http.NewInterface(
		listenHTTP,
		controller,
		gateway)

	ifaceHTTP.Start()
}