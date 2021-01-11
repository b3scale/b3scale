package main

import (
	"flag"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/config"
	"gitlab.com/infra.run/public/b3scale/pkg/events"
	"gitlab.com/infra.run/public/b3scale/pkg/logging"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

var version string = "HEAD"

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

func main() {
	fmt.Printf("b3scale node agent		v.%s\n", version)

	// Check if the enviroment was configured, when not try to
	// load the environment from .env or from a sysconfig env file
	if chk := config.EnvOpt(config.EnvDbURL, "unconfigured"); chk == "unconfigured" {
		config.LoadEnv([]string{
			".env",
			"/etc/sysconfig/b3scale",
		})
	}

	// Get config from env
	bbbPropFile := config.EnvOpt(config.EnvBBBConfig, config.EnvBBBConfigDefault)
	dbConnStr := config.EnvOpt(config.EnvDbURL, config.EnvDbURLDefault)
	loglevel := config.EnvOpt(config.EnvLogLevel, config.EnvLogLevelDefault)

	// Configure logging
	if err := logging.Setup(&logging.Options{
		Level: loglevel,
	}); err != nil {
		panic(err)
	}

	// Parse flags
	flag.Parse()

	// Parse BBB config
	bbbConf, err := config.ReadPropertiesFile(bbbPropFile)
	if err != nil {
		log.Fatal().
			Err(err).Msg("could not read bbb config")
	}

	log.Info().Msg("booting b3scalenoded")
	// Initialize postgres connection
	pool, err := store.Connect(dbConnStr)
	if err != nil {
		log.Fatal().Err(err).Msg("database connection")
	}

	// Get backend state, register backend if missing
	backend, err := configToBackendState(pool, bbbConf)
	if err != nil {
		log.Fatal().Err(err).Msg("load backend state")
	}
	if backend == nil {
		if !autoregister {
			log.Fatal().
				Msg("the backend was not found, " +
					"consider using the autoregister option " +
					" -register (or -a)")
		}
		backend, err = configRegisterBackendState(pool, bbbConf)
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("registering the backend failed")
		}
	}

	// Make redis client
	redisOpts, err := redis.ParseURL(configRedisURL(bbbConf))
	if err != nil {
		log.Fatal().Err(err).Msg("redis connection")
	}

	// Mark the presence of the noded
	go acquireBackendNodeLock(pool, backend)

	rdb := redis.NewClient(redisOpts)
	monitor := events.NewMonitor(rdb)
	channel := monitor.Subscribe()
	for ev := range channel {
		// We are handling an event in it's own goroutine
		go func(ev bbb.Event) {
			handler := NewEventHandler(pool, backend)
			err := handler.Dispatch(ev)
			if err != nil {
				log.Error().
					Err(err).
					Msg("event handler")
			}
		}(ev)
	}
}
