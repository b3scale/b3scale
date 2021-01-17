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

func main() {
	// Check if the enviroment was configured, when not try to
	// load the environment from .env or from a sysconfig env file
	if chk := config.EnvOpt(config.EnvDbURL, "unconfigured"); chk == "unconfigured" {
		config.LoadEnv([]string{
			".env",
			"/etc/sysconfig/b3scale",
		})
	}

	quit := make(chan bool)
	banner() // Most important.

	// Config
	listenHTTP := config.EnvOpt(config.EnvListenHTTP, config.EnvListenHTTPDefault)
	dbConnStr := config.EnvOpt(config.EnvDbURL, config.EnvDbURLDefault)
	loglevel := config.EnvOpt(config.EnvLogLevel, config.EnvLogLevelDefault)

	// Configure logging
	if err := logging.Setup(&logging.Options{
		Level: loglevel,
	}); err != nil {
		panic(err)
	}

	log.Info().Msg("booting b3scale")
	log.Debug().Str("url", dbConnStr).Msg("using database")

	// Initialize postgres connection
	dbConn, err := store.Connect(dbConnStr)
	if err != nil {
		log.Fatal().Err(err).Msg("database connection")
	}

	// Initialize cluster
	ctrl := cluster.NewController(dbConn)

	// Start router
	router := cluster.NewRouter(ctrl)
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

	go ifaceHTTP.Start(listenHTTP)

	<-quit
}
