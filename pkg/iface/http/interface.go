package http

import (
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/http2"

	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
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
	// Setup and configure echo framework
	e := echo.New()
	e.HideBanner = true

	// Middlewares
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} ${method} ${uri} ${status}, ${remote_ip}, ${latency_human}\n",
	}))
	e.Use(middleware.Recover())

	// We handle BBB requests in a custom middleware
	e.Use(BBBRequestMiddleware("/bbb", ctrl, gateway))

	// Prometheus Middleware
	// Find it under /metrics
	p := prometheus.NewPrometheus(serviceID, nil)
	p.Use(e)

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
	log.Println("Starting interface: HTTP")
	log.Fatal(iface.echo.Start(listen))
}

// StartCleartextHTTP2 starts a HTTP2 interface without
// any TLS encryption.
func (iface *Interface) StartCleartextHTTP2(listen string) {
	log.Println("Starting interface: HTTP2")
	s := &http2.Server{
		MaxConcurrentStreams: 200,
		MaxReadFrameSize:     1048576,
		IdleTimeout:          10 * time.Second,
	}
	log.Fatal(iface.echo.StartH2CServer(listen, s))
}

// Index / Root Handler
func (iface *Interface) httpIndex(c echo.Context) error {
	return c.HTML(
		http.StatusOK,
		"<h1>B3Scale! v0.1.0</h1>")
}
