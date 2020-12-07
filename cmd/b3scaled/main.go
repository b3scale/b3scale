package main

import (
	"github.com/rs/zerolog/log"

	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
	"gitlab.com/infra.run/public/b3scale/pkg/config"
	"gitlab.com/infra.run/public/b3scale/pkg/iface/http"
	"gitlab.com/infra.run/public/b3scale/pkg/logging"
	"gitlab.com/infra.run/public/b3scale/pkg/middlewares/routing"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

var version = "HEAD"

func main() {
	quit := make(chan bool)
	banner() // Most important.

	// Config
	listenHTTP := config.EnvOpt(
		"B3SCALE_LISTEN_HTTP", "127.0.0.1:42353") // B3S
	listenHTTP2 := config.EnvOpt(
		"B3SCALE_LISTEN_HTTP2", "127.0.0.1:42352") // B3S @ http2
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

	// Initialize cluster
	ctrl := cluster.NewController(dbConn)

	// Start router
	router := cluster.NewRouter(ctrl)
	router.Use(routing.Lookup(ctrl))
	router.Use(routing.SortLoad)

	// Start cluster request handler, and apply middlewares.
	// The middlewares are executes in reverse order.
	gateway := cluster.NewGateway(ctrl)
	// gateway.Use(requests.NewDispatchMerge()) // should not be required anymore
	gateway.Use(router.Middleware())

	// Start cluster controller
	go ctrl.Start()

	// Start HTTP interface
	ifaceHTTP := http.NewInterface("http", ctrl, gateway)
	ifaceHTTP2 := http.NewInterface("http2", ctrl, gateway)

	go ifaceHTTP.Start(listenHTTP)
	go ifaceHTTP2.StartCleartextHTTP2(listenHTTP2)

	<-quit
}
