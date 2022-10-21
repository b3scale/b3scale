package logging

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Middleware creates a logging middleware for echo
// using zerolog. This replace `lecho`, which suddenly broke.
func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			t0 := time.Now()

			// Invoke handler and log result
			handlerErr := next(c)
			duration := time.Since(t0)

			req := c.Request()
			res := c.Response()

			var event *zerolog.Event
			if handlerErr != nil {
				event = log.Error().Err(handlerErr)
			} else {
				event = log.Info()
			}

			event.Str("remote_ip", c.RealIP())
			event.Str("host", req.Host)
			event.Str("method", req.Method)
			event.Str("uri", req.RequestURI)
			event.Str("user_agent", req.UserAgent())
			event.Int("status", res.Status)
			event.Str("referer", req.Referer())
			event.Dur("latency", duration)
			event.Str("latency_human", duration.String())

			// Content size is only known for non error responses,
			// as the error response is generated by the error handler
			if handlerErr == nil {
				bytesIn, _ := strconv.ParseInt(
					req.Header.Get(echo.HeaderContentLength), 10, 64)
				bytesOut := res.Size
				event.Int64("bytes_in", bytesIn)
				event.Int64("bytes_out", bytesOut)
			}

			event.Send()

			return handlerErr
		}
	}
}