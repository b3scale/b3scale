package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"

	"github.com/b3scale/b3scale/pkg/config"
	"github.com/b3scale/b3scale/pkg/http/api"
	"github.com/b3scale/b3scale/pkg/http/api/client"
	"github.com/b3scale/b3scale/pkg/logging"
)

// Flags and parameters
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
	bbbPropFile := config.EnvOpt(
		config.EnvBBBConfig,
		config.EnvBBBConfigDefault,
	)
	// Parse BBB config
	props, err := config.ReadPropertiesFile(bbbPropFile)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("env", config.EnvBBBConfig).
			Str("file", bbbPropFile).
			Msg("could not read bbb config")
	}
	return props
}

func main() {
	ctx := context.Background()
	done := make(chan bool)

	fmt.Printf("b3scale node agent		v.%s\n", config.Version)

	flag.Parse()

	// Initialisation
	initEnvironment()
	initLogging()

	bbbCfg := initBBBConfig()
	b3s := initAPI()

	// Make sure we can talk to the API
	status, err := b3s.Status(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("api initialization failed")
	}
	log.Info().
		Str("agent", status.AccountRef).
		Str("api_version", status.API).
		Str("build", status.Build).
		Str("version", status.Version).
		Msg("connected to b3scaled")

	// Get registered backend
	backendCfg, err := backendFromConfig(bbbCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("backend from config")
	}

	backend, err := b3s.AgentBackendRetrieve(ctx)
	if err != nil && !errors.Is(err, api.ErrNotFound) {
		log.Fatal().Err(err).Msg("failed to get backend")
	}

	if backend == nil && !autoregister {
		log.Fatal().
			Msg("the backend was not found, " +
				"consider using the autoregister option " +
				" -register (or -a)")
	}

	// Set backend params from config
	if backend == nil {
		backend, err = b3s.BackendCreate(ctx, backendCfg)
		if err != nil {
			log.Fatal().Err(err).
				Msg("could not register backend")
		}
		log.Info().
			Str("id", backend.ID).
			Msg("registered backend")
	} else {
		update, err := json.Marshal(map[string]interface{}{
			"backend":     backendCfg.Backend,
			"load_factor": config.GetLoadFactor(),
		})
		if err != nil {
			log.Fatal().Err(err).
				Msg("could create backend update")
		}
		backend, err = b3s.BackendUpdateRaw(ctx, backend.ID, update)
		if err != nil {
			log.Fatal().Err(err).
				Msg("could not update backend")
		}
	}

	log.Info().
		Str("id", backend.ID).
		Str("host", backend.Backend.Host).
		Msg("agent configured for backend")

	redisOpts, err := redis.ParseURL(configRedisURL(bbbCfg))
	if err != nil {
		log.Fatal().Err(err).Msg("redis connection")
	}
	rdb := redis.NewClient(redisOpts)

	// Start heartbeat and monitoring
	go StartHeartbeat(ctx, b3s)
	go StartEventMonitor(ctx, b3s, rdb, backend)

	<-done
}
