package main

import (
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"

	"gitlab.com/infra.run/public/b3scale/pkg/config"
	"gitlab.com/infra.run/public/b3scale/pkg/events"
	"gitlab.com/infra.run/public/b3scale/pkg/logging"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

var version string = "HEAD"

func main() {
	fmt.Printf("b3scale node agent		v.%s\n", version)
	redisURL := config.EnvOpt(
		"BBB_REDIS_URL", "redis://localhost:6379/1")
	dbConnStr := config.EnvOpt(
		"B3SCALE_DB_URL",
		"postgres://postgres:postgres@localhost:5432/b3scale")
	loglevel := config.EnvOpt(
		"B3SCALE_LOG_LEVEL",
		"info")

	// Configure logging
	if err := logging.Setup(&logging.Options{
		Level: loglevel,
	}); err != nil {
		panic(err)
	}

	log.Info().Msg("booting b3scale")
	log.Info().Str("url", dbConnStr).Msg("using database")
	// Initialize postgres connection
	dbConn, err := store.Connect(dbConnStr)
	if err != nil {
		log.Fatal().Err(err).Msg("database connection")
	}

	// Make redis client
	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("redis connection")
	}

	rdb := redis.NewClient(redisOpts)
	monitor := events.NewMonitor(rdb)
	handler := NewEventHandler(dbConn)
	channel := monitor.Subscribe()
	for ev := range channel {
		err := handler.Dispatch(ev)
		if err != nil {
			log.Error().
				Err(err).
				Msg("event handler")
		}
	}
}
