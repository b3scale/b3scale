package main

import (
	"strconv"

	"github.com/rs/zerolog/log"

	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
	"gitlab.com/infra.run/public/b3scale/pkg/config"
	"gitlab.com/infra.run/public/b3scale/pkg/iface/http"
	"gitlab.com/infra.run/public/b3scale/pkg/logging"
	"gitlab.com/infra.run/public/b3scale/pkg/middlewares/requests"
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
	dbPoolSizeStr := config.EnvOpt(config.EnvDbPoolSize, config.EnvDbPoolSizeDefault)
	loglevel := config.EnvOpt(config.EnvLogLevel, config.EnvLogLevelDefault)
	revProxyEnabled := config.IsEnabled(config.EnvOpt(
		config.EnvReverseProxy, config.EnvReverseProxyDefault))

	dbPoolSize, err := strconv.Atoi(dbPoolSizeStr)

	// Configure logging
	if err := logging.Setup(&logging.Options{
		Level: loglevel,
	}); err != nil {
		panic(err)
	}

	log.Info().Msg("booting b3scale")
	log.Debug().Str("url", dbConnStr).Msg("using database")

	if revProxyEnabled {
		log.Info().Msg("reverse proxy mode is enabled")
	}

	// Initialize postgres connection
	err = store.Connect(&store.ConnectOpts{
		URL:      dbConnStr,
		MaxConns: int32(dbPoolSize),
		MinConns: 8,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("database connection")
	}

	log.Info().
		Int("maxConnections", dbPoolSize).
		Msg("database pool")

	// Initialize cluster
	ctrl := cluster.NewController()

	// Start router
	router := cluster.NewRouter(ctrl)
	router.Use(routing.SortLoad)

	// Start cluster request handler, and apply middlewares.
	// The middlewares are executes in reverse order.
	gateway := cluster.NewGateway(ctrl, &cluster.GatewayOptions{
		IsReverseProxyEnabled: revProxyEnabled,
	})

	gateway.Use(router.Middleware())
	gateway.Use(requests.DefaultPresentation())
	gateway.Use(requests.RewriteUniqueMeetingID())

	// Start cluster controller
	go ctrl.Start()

	// Start HTTP interface
	httpServer := http.NewServer("http", ctrl, gateway)
	go httpServer.Start(listenHTTP)

	<-quit
}
