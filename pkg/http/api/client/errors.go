package client

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/b3scale/b3scale/pkg/http/api"
)

// ErrNotFound is the error when a response is a 404
var ErrNotFound = errors.New("the resource could not be found (404)")

// ErrRequestFailed creates an APIError from the
// HTTP response.
func ErrRequestFailed(res *http.Response) api.ServerError {
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return api.ServerError{"error": err.Error()}
	}
	apiErr := api.ServerError{}
	if err := json.Unmarshal(body, &apiErr); err != nil {
		return api.ServerError{"error": body}
	}
	return apiErr
}
