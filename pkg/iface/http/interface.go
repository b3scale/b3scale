package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"github.com/ziflex/lecho/v2"

	"golang.org/x/net/http2"

	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
	"gitlab.com/infra.run/public/b3scale/pkg/config"
)

// Interface provides the http server for the
// application.
type Interface struct {
	serviceID  string
	echo       *echo.Echo
	gateway    *cluster.Gateway
	controller *cluster.Controller
}

// NewInterface configures and creates a new
// http interface to our cluster gateway.
func NewInterface(
	serviceID string,
	ctrl *cluster.Controller,
	gateway *cluster.Gateway,
) *Interface {
	logger := lecho.From(log.Logger)

	// Setup and configure echo framework
	e := echo.New()
	e.HideBanner = true

	// Middleware order: The middlewares are executed
	// in order of Use.
	e.Use(middleware.Recover())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 45 * time.Second,
	}))
	e.Use(lecho.Middleware(lecho.Config{
		Logger: logger,
	}))

	// Prometheus Middleware - Find it under /metrics
	p := prometheus.NewPrometheus(serviceID, nil)
	p.Use(e)

	// We handle BBB requests in a custom middleware
	e.Use(BBBRequestMiddleware("/bbb", ctrl, gateway))

	iface := &Interface{
		echo:       e,
		gateway:    gateway,
		controller: ctrl,
	}

	// Register index route
	e.GET("/", iface.httpIndex)

	return iface
}

// Start the HTTP interface
func (iface *Interface) Start(listen string) {
	log.Info().Msg("starting interface: HTTP")
	log.Fatal().
		Err(iface.echo.Start(listen)).
		Msg("starting http server")
}

// StartCleartextHTTP2 starts a HTTP2 interface without
// any TLS encryption.
func (iface *Interface) StartCleartextHTTP2(listen string) {
	log.Info().Msg("starting interface: PlaintextHTTP2")
	s := &http2.Server{
		MaxConcurrentStreams: 200,
		MaxReadFrameSize:     1048576,
		IdleTimeout:          10 * time.Second,
	}
	log.Fatal().
		Err(iface.echo.StartH2CServer(listen, s)).
		Msg("starting plaintext http2 server")

}

// Index / Root Handler
func (iface *Interface) httpIndex(c echo.Context) error {
	return c.HTML(
		http.StatusOK,
		fmt.Sprintf(
			"<h1>B3Scale! v.%s (%s)</h1>",
			config.Version, config.Build))
}
