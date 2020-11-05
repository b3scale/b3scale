package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	netHTTP "net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// BBBRequestMiddleware decodes the incoming HTTP request
// into a BBB request and passes it to the API gateway.
// All requests starting with the mountpoint prefix are
// treated as BBB requests.
//
// An error will be returned when A request can not be
// decoded.
func BBBRequestMiddleware(
	mountPoint string,
	ctrl *cluster.Controller,
) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Path()
			if !strings.HasPrefix(path, mountPoint) {
				return next(c) // nothing to do here.
			}
			// Decode HTTP request into a BBB request
			// and verify it.
			path = path[len(mountPoint):]
			frontendKey, action := decodePath(path)
			frontend, err := ctrl.GetFrontend(
				store.NewQuery().Eq("key", frontendKey))
			if err != nil {
				return handleAPIError(c, err)
			}
			if frontend == nil {
				return handleAPIError(c, fmt.Errorf(
					"no such frontend for key: %s", frontendKey))
			}

			// We have an action, we have a frontend, now
			// we need the query parameters and request body.
			body := readRequestBody(c)

			fmt.Println(action)
			fmt.Println(frontend)

			// Authenticate request

			// Decode BBB request
			fmt.Println(path)
			fmt.Println("query", c.QueryParams())

			return c.String(netHTTP.StatusOK, "huhuuuuu")
		}
	}
}

// decodePath extracts the frontend key and BBB
// action from the request path
func decodePath(path string) (string, string) {
	tokens := strings.Split(path, "/")
	if len(tokens) < 3 {
		return "", ""
	}
	return tokens[1], tokens[len(tokens)-1]
}

// handleAPIError is the error handler function
// for all API errors. The error will be wrapped into
// a BBB error response.
func handleAPIError(c echo.Context, err error) error {
	// Encode as BBB error
	res := &bbb.XMLResponse{
		Returncode: "ERROR",
		Message:    fmt.Sprintf("%s", err),
		MessageKey: "b3scale_server_error",
	}

	// Write error response
	return c.XML(netHTTP.StatusInternalServerError, res)
}

// readRequestBody will load the entire request body.
func readRequestBody(c echo.Context) []byte {
	body := []byte{}
	if c.Request().Body != nil { // Read
		body, _ = ioutil.ReadAll(c.Request().Body)
	}
	// Reset after reading
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return body
}
