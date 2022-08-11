package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/config"
	"github.com/b3scale/b3scale/pkg/events"
	"github.com/b3scale/b3scale/pkg/logging"
	"github.com/b3scale/b3scale/pkg/store"
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

func heartbeat(backend *store.BackendState) {
	ctx := context.Background()
	for {
		conn, err := store.Acquire(ctx)
		if err != nil {
			log.Error().Err(err).Msg("could not get heartbeat connection")
			time.Sleep(1 * time.Second)
			continue
		}
		tx, err := conn.Begin(ctx)
		if err != nil {
			log.Error().Err(err).Msg("could not start heartbeat tx")
		}

		if _, err := backend.UpdateAgentHeartbeat(ctx, tx); err != nil {
			log.Error().
				Err(err).
				Msg("update heartbeat failed")
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}

		conn.Release()

		time.Sleep(1 * time.Second)
	}
}

func main() {
	ctx := context.Background()

	fmt.Printf("b3scale node agent		v.%s\n", config.Version)

	// Check if the enviroment was configured, when not try to
	// load the environment from .env or from a sysconfig env file
	if chk := config.EnvOpt(config.EnvDbURL, "unconfigured"); chk == "unconfigured" {
		config.LoadEnv([]string{
			".env",
			"/etc/default/b3scale",
			"/etc/sysconfig/b3scale",
		})
	}

	// Get config from env
	bbbPropFile := config.EnvOpt(config.EnvBBBConfig, config.EnvBBBConfigDefault)
	dbConnStr := config.EnvOpt(config.EnvDbURL, config.EnvDbURLDefault)
	loglevel := config.EnvOpt(config.EnvLogLevel, config.EnvLogLevelDefault)
	loadFactor := config.GetLoadFactor()

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
	if err := store.Connect(&store.ConnectOpts{
		URL:      dbConnStr,
		MaxConns: 16,
		MinConns: 1,
	}); err != nil {
		log.Fatal().Err(err).Msg("database connection")
	}

	// Get backend state, register backend if missing
	backend, err := configToBackendState(ctx, bbbConf)
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
		backend, err = configRegisterBackendState(ctx, bbbConf)
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("registering the backend failed")
		}
	}

	// Set backend load factor
	backend.LoadFactor = loadFactor

	conn, err := store.Acquire(ctx)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("could not get connection")
	}
	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("could not get transaction")
	}
	if err := backend.Save(ctx, tx); err != nil {
		log.Fatal().
			Err(err).
			Msg("could not set backend load factor")
	}
	log.Info().
		Float64("loadFactor", loadFactor).
		Msg("setting load_factor")
	if err := tx.Commit(ctx); err != nil {
		log.Error().
			Err(err).
			Msg("setting load_factor")
	}
	conn.Release()

	// Make redis client
	redisOpts, err := redis.ParseURL(configRedisURL(bbbConf))
	if err != nil {
		log.Fatal().Err(err).Msg("redis connection")
	}

	// Mark the presence of the noded
	go heartbeat(backend)

	rdb := redis.NewClient(redisOpts)
	monitor := events.NewMonitor(rdb)
	channel := monitor.Subscribe()
	for ev := range channel {
		// We are handling an event in it's own goroutine
		go func(ev bbb.Event) {
			handler := NewEventHandler(backend)
			ctx, cancel := context.WithTimeout(
				context.Background(), 15*time.Second)
			defer cancel()

			err := handler.Dispatch(ctx, ev)
			if err != nil {
				log.Error().
					Err(err).
					Msg("event handler")
			}
		}(ev)
	}
}
