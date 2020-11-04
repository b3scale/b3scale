package http

import (
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
)

// BBBRequestMiddleware decodes the incoming HTTP request
// into a BBB request and passes it to the API gateway.
// All requests starting with the mountpoint prefix are
// treated as BBB requests.
//
// An error will be returned when A request can not be
// decoded.
func BBBRequestMiddleware(mountPoint string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Path()
			if !strings.HasPrefix(path, mountPoint) {
				return next(c) // nothing to do here.
			}
			// Strip prefix
			path = path[len(mountPoint):]

			// frontendKey, action := decodePath(path)

			// Authenticate frontend

			// Decode BBB request
			fmt.Println(path)
			fmt.Println("query", c.QueryParams())

			return next(c)
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
