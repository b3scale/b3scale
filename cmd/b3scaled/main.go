package main

import (
	"log"
	"time"

	"gitlab.com/infra.run/public/b3scale/mod/cluster"
	"gitlab.com/infra.run/public/b3scale/mod/config"
)

func main() {
	banner() // Most important.
	log.Println("Starting b3scaled.")

	// Parse flags
	frontendsConfigFilename := "etc/b3scale/frontends.conf"
	backendsConfigFilename := "etc/b3scale/backends.conf"

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
