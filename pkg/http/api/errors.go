package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/b3scale/b3scale/pkg/store"
	"github.com/b3scale/b3scale/pkg/templates"
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

// ServerError will contain the decoded json body
// when the response status was not OK or Accepted
type ServerError map[string]interface{}

// Error implements the error interface
func (err ServerError) Error() string {
	errs := []string{}
	for k, v := range err {
		errs = append(errs, fmt.Sprintf("%s: %v", k, v))
	}
	return strings.Join(errs, "; ")
}

// ErrNotFound is the error when a response is a 404
var ErrNotFound = errors.New("the resource could not be found (404)")

// HTMLError will render an HTML error page
func HTMLError(c echo.Context, status int, title, message string) error {
	body := templates.ErrorPage(title, message)
	return c.HTMLBlob(status, body)
}
