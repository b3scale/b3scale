package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	netHTTP "net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

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
	gateway *cluster.Gateway,
) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			tx, err := ctrl.BeginTx(ctx)
			if err != nil {
				return err
			}
			defer tx.Rollback(ctx)
			ctx = store.ContextWithTransaction(ctx, tx)

			path := c.Path()
			if !strings.HasPrefix(path, mountPoint) {
				return next(c) // nothing to do here.
			}
			// Decode HTTP request into a BBB request
			// and verify it.
			path = path[len(mountPoint):]
			frontendKey, resource := decodePath(path)
			frontend, err := ctrl.GetFrontend(ctx, store.Q().
				Where("key = ?", frontendKey))
			if err != nil {
				return handleAPIError(c, err)
			}
			if frontend == nil {
				return handleAPIError(c, fmt.Errorf(
					"no such frontend for key: %s", frontendKey))
			}

			// We have an action, we have a frontend, now
			// we need the query parameters and request body.
			params := decodeParams(c)
			checksum, _ := params.Checksum()
			body := readRequestBody(c)

			bbbReq := &bbb.Request{
				Request:  c.Request(),
				Frontend: frontend.Frontend(),
				Resource: resource,
				Params:   params,
				Body:     body,
				Checksum: checksum,
			}

			// Authenticate request
			if err := bbbReq.Verify(); err != nil {
				return handleAPIError(c, err)
			}

			res := gateway.Dispatch(ctx, bbbReq)

			if err := tx.Commit(ctx); err != nil {
				log.Error().
					Err(err).
					Msg("tx commit failed after request")
			}

			return writeBBBResponse(c, res)
		}
	}
}

// writeBBBResponse takes a response from the cluster
// and writes it as a response to the request.
func writeBBBResponse(c echo.Context, res bbb.Response) error {
	// When the status is not set assume something went wrong
	status := res.Status()
	if status == 0 {
		status = netHTTP.StatusInternalServerError
	}

	// Update and write headers
	for key, values := range res.Header() {
		for _, v := range values {
			c.Response().Header().Add(key, v)
		}
	}
	c.Response().WriteHeader(status)

	// Serialize BBB response and send
	data, err := res.Marshal()
	if err != nil {
		return err
	}
	_, err = c.Response().Write(data)
	return err
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

func decodeParams(c echo.Context) bbb.Params {
	values := c.QueryParams()
	params := bbb.Params{}
	for k := range values {
		params[k] = values.Get(k)
	}
	return params
}
