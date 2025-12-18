package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	pclient "github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/cluster"
	"github.com/b3scale/b3scale/pkg/config"
	"github.com/b3scale/b3scale/pkg/http/api"
	"github.com/b3scale/b3scale/pkg/http/static"
	"github.com/b3scale/b3scale/pkg/logging"
	"github.com/b3scale/b3scale/pkg/metrics"
	"github.com/b3scale/b3scale/pkg/templates"
)

// Server provides the http server for the application.
type Server struct {
	serviceID  string
	echo       *echo.Echo
	gateway    *cluster.Gateway
	controller *cluster.Controller
}

// NewServer configures and creates a new http interface
// to our cluster gateway.
func NewServer(
	serviceID string,
	ctrl *cluster.Controller,
	gateway *cluster.Gateway,
) *Server {
	// Setup and configure echo framework
	e := echo.New()
	e.HideBanner = true

	// Middleware order: The middlewares are executed
	// in order of Use.
	e.Use(middleware.Recover())
	e.Use(logging.Middleware())
	e.Use(middleware.ContextTimeoutWithConfig(middleware.ContextTimeoutConfig{
		Timeout: config.GetHTTPRequestTimeout(),
	}))

	// Prometheus Middleware - Find it under /metrics
	e.Use(echoprometheus.NewMiddleware(serviceID))
	e.GET("/metrics", echoprometheus.NewHandler())

	pclient.MustRegister(metrics.Collector{})

	// We handle BBB requests in a custom middleware
	e.Use(BBBRequestMiddleware("/bbb", ctrl, gateway))

	s := &Server{
		serviceID:  serviceID,
		echo:       e,
		gateway:    gateway,
		controller: ctrl,
	}

	// Register routes
	e.GET("/", s.httpIndex)
	e.GET("/docs/api/v1", s.apiDocsShow)
	e.GET("/static/*", echo.WrapHandler(static.AssetsHTTPHandler("/static")))
	e.GET("/b3s/retry-join/:req", s.httpRetryJoin)

	if err := api.Init(e); err != nil {
		log.Warn().Err(err).Msg("could not initialize rest API")
	}

	return s
}

// Start the HTTP interface
func (s *Server) Start(listen string) {
	log.Info().Msg("starting interface: HTTP")
	httpServer := &http.Server{
		Addr:              listen,
		ReadHeaderTimeout: config.GetHTTPReadHeaderTimeout(),
		WriteTimeout:      config.GetHTTPWriteTimeout(),
		IdleTimeout:       config.GetHTTPIdleTimeout(),
	}

	log.Fatal().
		Err(s.echo.StartServer(httpServer)).
		Msg("starting http server")
}

// Index / Root Handler
func (s *Server) httpIndex(c echo.Context) error {
	return c.HTML(
		http.StatusOK,
		fmt.Sprintf(
			"<h1>B3Scale! v.%s (%s)</h1>",
			config.Version, config.Build))
}

func (s *Server) apiDocsShow(c echo.Context) error {
	return c.Redirect(http.StatusFound, "/static/docs/api-v1.html")
}

// Internal / Retry Join Handler
func (s *Server) httpRetryJoin(c echo.Context) error {
	// Restore join URL from request
	// Please note, that the blob is made opaque because it
	// contains information like a password etc which could
	// irritate users. HOWEVER: this information is not really
	// a secret, as it is done clientsided when klicking "join".
	//
	req, err := bbb.UnmarshalURLSafeRequest([]byte(c.Param("req")))
	if err != nil {
		return err
	}
	joinURL := req.Request.URL.String()

	// Prevent open redirects. This is already a slippery slope.
	if !strings.HasPrefix(joinURL, "/bbb") {
		return fmt.Errorf("invalid join URL")
	}

	body := templates.RetryJoin(joinURL)
	return c.HTMLBlob(http.StatusOK, body)
}
