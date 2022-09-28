package main

import (
	"context"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/b3scale/b3scale/pkg/config"
	"github.com/b3scale/b3scale/pkg/logging"
)

func main() {
	// Check if the enviroment was configured, when not try to
	// load the environment from .env or from a sysconfig env file
	chkEnv := config.EnvOpt(config.EnvDbURL, "unconfigured")
	if chkEnv == "unconfigured" {
		// Try to load the environment from files
		config.LoadEnv([]string{
			".env",
			"/etc/default/b3scale",
			"/etc/sysconfig/b3scale",
		})
	}

	loglevel := config.EnvOpt(config.EnvLogLevel, config.EnvLogLevelDefault)
	if err := logging.Setup(&logging.Options{
		Level: loglevel,
	}); err != nil {
		panic(err)
	}

	// Start the CLI
	cli := NewCli()
	if err := cli.Run(context.Background(), os.Args); err != nil {
		log.Fatal().Err(err).Msg("exit with fatal error")
	}

	// A note about the return code:
	// Operations will usually return a success (0)
	// or an error 1. However, we distinguish between
	// a successful operation (0) and an operation, which
	// was not applied because there was no change (64)
	os.Exit(cli.returnCode)
}
