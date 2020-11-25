package main

import (
	"log"
	"os"

	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

var version = "HEAD"

func getopt(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func main() {
	dbConnStr := getopt(
		"B3SCALE_DB_URL",
		"postgres://postgres:postgres@localhost:5432/b3scale")
	dbPool, err := store.Connect(dbConnStr)
	if err != nil {
		log.Fatal(err)
	}
	queue := store.NewCommandQueue(dbPool)

	// Start the CLI
	cli := NewCli(queue, dbPool)
	if err := cli.Run(os.Args); err != nil {
		log.Fatal(err)
	}

	// A note about the return code:
	// Operations will usually return a success (0)
	// or an error 1. However, we distinguish between
	// a successful operation (0) and an operation, which
	// was not applied because there was no change (64)
	os.Exit(cli.returnCode)
}
