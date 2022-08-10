package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/b3scale/b3scale/pkg/store"
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

// ErrorValidationFailed creates an API response when validating
// a resource failed.
func ErrorValidationFailed(c echo.Context, err store.ValidationError) error {
	return c.JSON(http.StatusBadRequest, map[string]interface{}{
		"error":   "validation_error",
		"message": "validation failed with input",
		"fields":  err,
	})
}

// ErrorHandler intercepts well known errors
// and renders a response.
func ErrorHandler(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)
		if err == nil {
			return nil
		}
		if validationErr, ok := err.(store.ValidationError); ok {
			return ErrorValidationFailed(c, validationErr)
		}
		return err
	}
}
