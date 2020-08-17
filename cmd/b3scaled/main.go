package main

import (
	"log"
	"os"
	"time"

	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
	"gitlab.com/infra.run/public/b3scale/pkg/config"
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

	// Parse flags
	frontendsConfigFilename := getopt(
		"B3SCALE_FRONTENDS", "etc/b3scale/frontends.conf")
	backendsConfigFilename := getopt(
		"B3SCALE_BACKENDS", "etc/b3scale/backends.conf")

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

	// Start cluster router
	router := cluster.NewRouter(controller)
	go router.Start()

	// Start ctrl interface

	// Start HTTP interface

	// Just for testing...
	time.Sleep(1 * time.Second)
}
