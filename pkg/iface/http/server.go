package http

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	pclient "github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"github.com/ziflex/lecho/v2"
	"golang.org/x/net/http2"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
	"gitlab.com/infra.run/public/b3scale/pkg/config"
	"gitlab.com/infra.run/public/b3scale/pkg/metrics"
	"gitlab.com/infra.run/public/b3scale/pkg/templates"
)

const (
	// RequestTimeout until the request has to be finished
	RequestTimeout = 60 * time.Second
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
	logger := lecho.From(log.Logger)

	// Setup and configure echo framework
	e := echo.New()
	e.HideBanner = true

	// Middleware order: The middlewares are executed
	// in order of Use.
	e.Use(middleware.Recover())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: RequestTimeout,
	}))
	e.Use(lecho.Middleware(lecho.Config{
		Logger: logger,
	}))

	// Prometheus Middleware - Find it under /metrics
	p := prometheus.NewPrometheus(serviceID, nil)
	p.Use(e)

	pclient.MustRegister(metrics.Collector{})

	// We handle BBB requests in a custom middleware
	e.Use(BBBRequestMiddleware("/bbb", ctrl, gateway))

	s := &Server{
		echo:       e,
		gateway:    gateway,
		controller: ctrl,
	}

	// Register routes
	e.GET("/", s.httpIndex)
	e.GET("/b3s/retry-join/:req", s.httpRetryJoin)

	return s
}

// Start the HTTP interface
func (s *Server) Start(listen string) {
	log.Info().Msg("starting interface: HTTP")
	httpServer := &http.Server{
		Addr:              listen,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	log.Fatal().
		Err(s.echo.StartServer(httpServer)).
		Msg("starting http server")
}

// StartCleartextHTTP2 starts a HTTP2 interface without
// any TLS encryption.
func (s *Server) StartCleartextHTTP2(listen string) {
	log.Info().Msg("starting interface: CleartextHTTP2")
	httpServer := &http2.Server{
		MaxConcurrentStreams: 200,
		MaxReadFrameSize:     1048576,
		IdleTimeout:          10 * time.Second,
	}
	log.Fatal().
		Err(s.echo.StartH2CServer(listen, httpServer)).
		Msg("starting plaintext http2 server")

}

// Index / Root Handler
func (s *Server) httpIndex(c echo.Context) error {
	return c.HTML(
		http.StatusOK,
		fmt.Sprintf(
			"<h1>B3Scale! v.%s (%s)</h1>",
			config.Version, config.Build))
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
