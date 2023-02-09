package http

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	netHTTP "net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	//	"github.com/rs/zerolog/log"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/cluster"
	"github.com/b3scale/b3scale/pkg/store"
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
			// We use the request context. However, for some things
			// we need to make sure they are persisted even though
			// the context is canceled. This might happen when the
			// client disconnects after we made our request to the
			// backend.
			// ctx := c.Request().Context()

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			path := c.Path()
			if !strings.HasPrefix(path, mountPoint) {
				return next(c) // nothing to do here.
			}

			// We acquire a connection to the database here,
			// if this fails it does not really make sense to move on.
			// TODO: See if we actually can use this context.
			conn, err := store.Acquire(ctx)
			if err != nil {
				return err
			}
			defer conn.Release()
			ctx = store.ContextWithConnection(ctx, conn)

			// Decode HTTP request into a BBB request
			// and verify it.
			path = path[len(mountPoint):]
			frontendKey, resource := decodePath(path)

			frontend, err := cluster.GetFrontend(ctx, store.Q().
				Where("key = ?", frontendKey))
			if err != nil {
				return handleAPIError(c, err)
			}

			// Check if the frontend could be identified and it is not disabled
			if frontend == nil || (frontend != nil && frontend.Active() == false) {
				return handleAPIError(c, fmt.Errorf(
					"no such frontend for key: %s", frontendKey))
			}
			ctx = cluster.ContextWithFrontend(ctx, frontend)

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

			log.Debug().Stringer("req", bbbReq).Msg("inbound request")

			// Authenticate request
			if err := bbbReq.Verify(); err != nil {
				return handleAPIError(c, err)
			}

			// Before we dispatch, let's check if the original
			// request context is still valid
			if err := c.Request().Context().Err(); err != nil {
				return err
			}

			res := gateway.Dispatch(ctx, conn, bbbReq)

			return writeBBBResponse(c, res)
		}
	}
}

// writeBBBResponse takes a response from the cluster
// and writes it as a response to the request.
func writeBBBResponse(c echo.Context, res bbb.Response) error {
	// Check if the context is still valid
	if err := c.Request().Context().Err(); err != nil {
		return err
	}

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
