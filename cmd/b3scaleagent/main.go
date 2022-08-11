package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/b3scale/b3scale/pkg/config"
	"github.com/b3scale/b3scale/pkg/http/api"
	"github.com/b3scale/b3scale/pkg/http/api/client"
	"github.com/b3scale/b3scale/pkg/logging"
)

// Flags and parameters, same as noded
var (
	autoregister bool
)

func init() {
	usage := "automatically register backend node in cluster"
	flag.BoolVar(
		&autoregister, "register", false, usage)
	flag.BoolVar(
		&autoregister, "a", false, usage+" (shorthand)")
}

// initEnvironment checks if the environment is configured
// or tries to load it from well known locations
func initEnvironment() {
	if _, ok := config.GetEnvOpt(config.EnvAPIURL); ok {
		return // nothing to do here
	}
	// load the environment from .env or from a sysconfig env file
	config.LoadEnv([]string{
		".env",
		"/etc/default/b3scale",
		"/etc/sysconfig/b3scale",
	})
}

// initLogging configures logging
func initLogging() {
	loglevel := config.EnvOpt(config.EnvLogLevel, config.EnvLogLevelDefault)
	if err := logging.Setup(&logging.Options{
		Level: loglevel,
	}); err != nil {
		panic(err)
	}
}

// initAPI initializes the API client
func initAPI() api.Client {
	// Configure client
	apiURL, ok := config.GetEnvOpt(config.EnvAPIURL)
	if !ok {
		log.Fatal().Msg(config.EnvAPIURL + " is not configured")
	}
	accessToken, ok := config.GetEnvOpt(config.EnvAPIAccessToken)
	if !ok {
		log.Fatal().Msg(config.EnvAPIAccessToken + " is not configured")
	}
	log.Info().Str("host", apiURL).Msg("api")
	return client.New(apiURL, accessToken)
}

func initBBBConfig() config.Properties {
	bbbPropFile := config.EnvOpt(config.EnvBBBConfig, config.EnvBBBConfigDefault)
	// Parse BBB config
	props, err := config.ReadPropertiesFile(bbbPropFile)
	if err != nil {
		log.Fatal().
			Err(err).
			String("file", bbbPropFile).
			Msg("could not read bbb config")
	}
	return props
}

func main() {
	ctx := context.Background()
	fmt.Printf("b3scale node agent		v.%s\n", config.Version)

	// Initialisation
	initEnvironment()
	initLogging()

	client := initAPI()

	status, err := client.Status(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("api initialization failed")
	}
	log.Info().
		Str("agent", status.AccountRef).
		Str("api_version", status.API).
		Str("build", status.Build).
		Str("version", status.Version).
		Msg("connected to b3scaled")

	/*
		loadFactor := config.GetLoadFactor()
	*/

	// Check API connection

}
