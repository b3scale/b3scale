package main

import (
	"log"
	"os"

	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

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
}
