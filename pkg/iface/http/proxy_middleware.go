package http

import (
	// "bytes"
	// "compress/gzip"
	// "io"
	// "io/ioutil"
	"net/http"
	"net/url"
	// "regexp"
	// "strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"

	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// BBBProxyMiddleware is a reverse proxy middleware
// which will forward requests to a bbb backend
func BBBProxyMiddleware(
	ctrl *cluster.Controller,
) echo.MiddlewareFunc {

	// The echo framework provides us with a proxy
	// middleware. We will invoke the middleware handler
	// and use our own proxy balancer
	proxyConfig := middleware.ProxyConfig{
		ContextKey: "target",
		Skipper: func(c echo.Context) bool {
			path := c.Path()

			if !strings.HasPrefix(path, "/html5client") &&
				!strings.HasPrefix(path, "/bigbluebutton") &&
				!strings.HasPrefix(path, "/ws") {
				return true // Nothing to do here
			}

			if c.IsWebSocket() {
				log.Info().Msg("!!!WEBSOCKET DETECTED!!!")
			}

			bidcookie, err := c.Cookie("B3SBID")
			if err != nil {
				log.Error().Err(err).Msg("backend identification cookie error")
				return true
			}

			backendID := bidcookie.Value
			backend, err := ctrl.GetBackend(store.Q().
				Where("id = ?", backendID))
			if err != nil {
				log.Error().
					Err(err).
					Msg("get backend for routing failed")
				return true
			}
			if backend == nil {
				log.Info().
					Str("backendID", backendID).
					Msg("could not find backend")
				return true
			}

			// Set backend for balancer
			c.Set("backend", backend)
			return false
		},
		// Transport: NewBBBProxyRewriteTransport(),
		Balancer: NewBBBProxyBalancer(ctrl),
	}

	return middleware.ProxyWithConfig(proxyConfig)
}

// BBBProxyRewriteTransport wraps a http default
// roundtripper and will modify the request and response
type BBBProxyRewriteTransport struct {
	http.RoundTripper
}

// NewBBBProxyRewriteTransport creates a new rewriter
func NewBBBProxyRewriteTransport() *BBBProxyRewriteTransport {
	return &BBBProxyRewriteTransport{
		http.DefaultTransport,
	}
}

// RoundTrip invokes the http transport and modifies
// the response.
func (t *BBBProxyRewriteTransport) RoundTrip(
	req *http.Request,
) (*http.Response, error) {
	backendID := req.Header.Get("X-B3Scale-BackendID")

	log.Info().
		Str("backendID", backendID).
		Str("url", req.URL.String()).
		Msg("PROXY REQUEST")

	res, err := t.RoundTripper.RoundTrip(req)
	/*

		// Rewrite redirects
		location := res.Header.Get("Location")
		if location != "" {
			locURL, _ := url.Parse(location)
			locURL.Scheme = ""
			locURL.Host = ""

			if locURL.Path == "" {
				locURL.Path = "/"
			}
			locURL.Path = "/client/" + backendID + locURL.Path

			res.Header.Set("Location", locURL.String())
		}

		// Wrap body reader
		contentEncoding := res.Header.Get("Content-Encoding")
		var reader io.ReadCloser
		switch contentEncoding {
		case "gzip":
			reader, _ = gzip.NewReader(res.Body)
			defer reader.Close()
		default:
			reader = res.Body
		}

		// Receive body data and rewrite urls if required
		body, err := ioutil.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		err = res.Body.Close()
		if err != nil {
			return nil, err
		}

		contentType := res.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "text") ||
			strings.HasPrefix(contentType, "application") {
			body = rewriteBodyURLs(body, backendID)
		}

		// Todo: Encode gziped
		res.Header.Del("Content-Encoding")
		res.Body = ioutil.NopCloser(bytes.NewReader(body))

		res.ContentLength = int64(len(body))
		res.Header.Set("Content-Length", strconv.Itoa(len(body)))
	*/

	return res, err
}

// Replace absolute urls with the /client/backendID
// prefix
func rewriteBodyURLs(body []byte, backendID string) []byte {
	if body == nil {
		log.Error().Msg("BODY WAS NIL")
		return nil
	}
	/*
			prefix := []byte("=\"/client/" + backendID + "/")
			return ReMatchAbsoluteURL.ReplaceAll(body, prefix)
		prefix := []byte("/client/" + backendID + "/html5client/")
		body = ReMatchHTML5Client.ReplaceAll(body, prefix)
	*/
	return body
}

// BBBProxyBalancer will select the backend
// from the clusted and create a proxy target
type BBBProxyBalancer struct {
	ctrl *cluster.Controller
}

// NewBBBProxyBalancer creates a new balancer
func NewBBBProxyBalancer(ctrl *cluster.Controller) *BBBProxyBalancer {
	return &BBBProxyBalancer{
		ctrl: ctrl,
	}
}

// AddTarget to the proxy balancer. We do not
// support this.
func (b *BBBProxyBalancer) AddTarget(*middleware.ProxyTarget) bool {
	return false
}

// RemoveTarget from the balancer. We also do
// not support this.
func (b *BBBProxyBalancer) RemoveTarget(string) bool {
	return false
}

// Next selects the proxy target. We identify
// the backend by ID and strip it from the the
// request via a rewrite rule.
func (b *BBBProxyBalancer) Next(c echo.Context) *middleware.ProxyTarget {
	backend := c.Get("backend").(*cluster.Backend)
	backendURL, _ := url.Parse(backend.Host())
	backendURL.Path = "/"
	target := &middleware.ProxyTarget{
		Name: backend.ID(),
		URL:  backendURL,
	}

	return target
}

// Remove the backendID from the path
// and get the target path.
func decodeClientProxyPath(prefix, path string) (string, string) {
	// Strip prefix and leading slash
	path = path[len(prefix):]
	if path == "" {
		return "", ""
	}
	if path[0] == '/' {
		path = path[1:]
	}

	t := strings.SplitN(path, "/", 2)
	if len(t) < 2 {
		return t[0], "/"
	}
	return t[0], "/" + t[1]
}
