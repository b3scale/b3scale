package main

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/b3scale/b3scale/pkg/cluster"
	"github.com/b3scale/b3scale/pkg/config"
	"github.com/b3scale/b3scale/pkg/http"
	"github.com/b3scale/b3scale/pkg/logging"
	"github.com/b3scale/b3scale/pkg/middlewares/requests"
	"github.com/b3scale/b3scale/pkg/middlewares/routing"
	"github.com/b3scale/b3scale/pkg/store"
)

func main() {
	ctx := context.Background()

	// Check if the enviroment was configured, when not try to
	// load the environment from .env or from a sysconfig env file
	if chk := config.EnvOpt(config.EnvDbURL, "unconfigured"); chk == "unconfigured" {
		config.LoadEnv([]string{
			".env",
			"/etc/default/b3scale",
			"/etc/sysconfig/b3scale",
		})
	}

	banner() // Most important.

	// Configure logging
	if err := logging.Setup(config.GetLoggingOpts()); err != nil {
		panic(err)
	}

	// Ensure all required configuration is present
	if err := config.CheckEnv(); err != nil {
		log.Fatal().Err(err).Msg("configuration incomplete")
		return
	}

	// Confiure database and server settings
	listenHTTP := config.EnvOpt(config.EnvListenHTTP, config.EnvListenHTTPDefault)

	// Begin server initialization
	log.Info().Msg("booting b3scale")

	revProxyEnabled := config.IsEnabled(config.EnvOpt(
		config.EnvReverseProxy, config.EnvReverseProxyDefault))
	if revProxyEnabled {
		log.Info().Msg("reverse proxy mode is enabled")
	}

	// Initialize postgres connection
	if err := store.Connect(config.GetDbConnectOpts()); err != nil {
		log.Fatal().Err(err).Msg("database connection")
	}

	// Recordings are an optional feature, so we will treat errors
	// as warnings for now.
	recordingsStorage, err := store.NewRecordingsStorageFromEnv()
	if err != nil {
		log.Error().Err(err).Msg("could not initialize recordings storage")
		log.Error().Msg("recording feature is not available")
	} else {
		if err := recordingsStorage.Check(); err != nil {
			log.Error().Err(err).Msg("recordings storage error")
		}
	}

	// Initialize cluster
	ctrl := cluster.NewController()

	// Create router and configure middlewares.
	// IMPORTANT: The middlewares are executed in reverse order.
	router := cluster.NewRouter(ctrl)
	router.Use(routing.SortLoad)
	router.Use(routing.RequiredTags)

	// Start cluster request handler, and apply middlewares.
	// IMPORTANT: The middlewares are executed in reverse order.
	gateway := cluster.NewGateway(ctrl, &cluster.GatewayOptions{})

	gateway.Use(requests.AdminRequestHandler(router))
	gateway.Use(requests.RecordingsRequestHandler(
		router, &requests.RecordingsHandlerOptions{}))
	gateway.Use(requests.MeetingsRequestHandler(
		router, &requests.MeetingsHandlerOptions{
			UseReverseProxy: revProxyEnabled,
		}))

	gateway.Use(requests.SetMetaFrontend())
	gateway.Use(requests.SetDefaultPresentation())
	gateway.Use(requests.SetCreateParams())
	gateway.Use(requests.CheckAttendeesLimit())
	gateway.Use(requests.BindMeetingFrontend())
	gateway.Use(requests.RewriteMetaCallbackURLs())
	gateway.Use(requests.RewriteUniqueMeetingID())

	// Start cluster controller
	go ctrl.Start(ctx)

	// Start HTTP interface
	httpServer := http.NewServer("http", ctrl, gateway)
	go httpServer.Start(listenHTTP)

	<-ctx.Done()
}
