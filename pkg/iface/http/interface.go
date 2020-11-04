package http

import (
	"log"
	"net/http"

	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
)

// Interface provides the http server for the
// application.
type Interface struct {
	listen     string
	echo       *echo.Echo
	gateway    *cluster.Gateway
	controller *cluster.Controller
}

// NewInterface configures and creates a new
// http interface to our cluster gateway.
func NewInterface(
	listen string,
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
	e.Use(BBBRequestMiddleware("/bbb"))

	// Prometheus Middleware
	// Find it under /metrics
	p := prometheus.NewPrometheus("echo", nil)
	p.Use(e)

	iface := &Interface{
		echo:       e,
		listen:     listen,
		gateway:    gateway,
		controller: ctrl,
	}

	// Register index route
	e.GET("/", iface.httpIndex)

	return iface
}

// Start the HTTP interface
func (iface *Interface) Start() {
	log.Println("Starting interface: HTTP")
	log.Fatal(iface.echo.Start(iface.listen))
}

// Index / Root Handler
func (iface *Interface) httpIndex(c echo.Context) error {
	return c.HTML(
		http.StatusOK,
		"<h1>B3Scale! v0.1.0</h1>")
}
