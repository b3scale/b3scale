package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/config"
	"gitlab.com/infra.run/public/b3scale/pkg/events"
	"gitlab.com/infra.run/public/b3scale/pkg/logging"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
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

type txFunc func(context.Context) error

func withTransactionContext(
	pool *pgxpool.Pool,
	timeout time.Duration,
	txFunc txFunc,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	ctx = store.ContextWithTransaction(ctx, tx)

	if err := txFunc(ctx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func heartbeat(pool *pgxpool.Pool, backend *store.BackendState) {
	timeout := 10 * time.Second
	for {
		if err := withTransactionContext(pool, timeout, func(ctx context.Context) error {
			return backend.UpdateAgentHeartbeat(ctx)
		}); err != nil {
			log.Error().
				Err(err).
				Msg("update heartbeat failed")
		}
		time.Sleep(1 * time.Second)
	}
}

func main() {
	fmt.Printf("b3scale node agent		v.%s\n", config.Version)

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
	pool, err := store.Connect(&store.ConnectOpts{
		URL:      dbConnStr,
		MaxConns: 16,
		MinConns: 1,
	})
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

	// Set backend load factor
	backend.LoadFactor = loadFactor
	if err := withTransactionContext(pool, 10*time.Second, func(ctx context.Context) error {
		return backend.Save(ctx)
	}); err != nil {
		log.Fatal().
			Err(err).
			Msg("could not set backend load factor")
	}
	log.Info().
		Float64("loadFactor", loadFactor).
		Msg("setting load_factor")

	// Make redis client
	redisOpts, err := redis.ParseURL(configRedisURL(bbbConf))
	if err != nil {
		log.Fatal().Err(err).Msg("redis connection")
	}

	// Mark the presence of the noded
	go heartbeat(pool, backend)

	rdb := redis.NewClient(redisOpts)
	monitor := events.NewMonitor(rdb)
	channel := monitor.Subscribe()
	for ev := range channel {
		// We are handling an event in it's own goroutine
		go func(ev bbb.Event) {
			handler := NewEventHandler(backend)
			err := withTransactionContext(
				pool, 10*time.Second, func(ctx context.Context) error {
					return handler.Dispatch(ctx, ev)
				})
			if err != nil {
				log.Error().
					Err(err).
					Msg("event handler")
			}
		}(ev)
	}
}
