package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// ErrorInvalidCredentials will create an API response
// for an unauthorized request
func ErrorInvalidCredentials(c echo.Context) error {
	return c.JSON(http.StatusForbidden, map[string]string{
		"error": "invalid_credentials",
		"message": "the credentials provided are lacking the " +
			"scope: b3scale or b3scale:admin",
	})
}
