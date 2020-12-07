package main

import (
	"os"

	"github.com/rs/zerolog/log"

	"gitlab.com/infra.run/public/b3scale/pkg/config"
	"gitlab.com/infra.run/public/b3scale/pkg/logging"
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
	dbConnStr := config.EnvOpt(
		"B3SCALE_DB_URL",
		"postgres://postgres:postgres@localhost:5432/b3scale")
	loglevel := config.EnvOpt(
		"B3SCALE_LOG_LEVEL",
		"info")

	if err := logging.Setup(&logging.Options{
		Level: loglevel,
	}); err != nil {
		panic(err)
	}

	dbPool, err := store.Connect(dbConnStr)
	if err != nil {
		log.Fatal().Err(err).Msg("database connection")
	}
	queue := store.NewCommandQueue(dbPool)

	// Start the CLI
	cli := NewCli(queue, dbPool)
	if err := cli.Run(os.Args); err != nil {
		log.Fatal().Err(err).Msg("this is fatal")
	}

	// A note about the return code:
	// Operations will usually return a success (0)
	// or an error 1. However, we distinguish between
	// a successful operation (0) and an operation, which
	// was not applied because there was no change (64)
	os.Exit(cli.returnCode)
}
