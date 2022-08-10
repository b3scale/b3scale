package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// APIError will contain the decoded json body
// when the response status was not OK or Accepted
type APIError map[string]interface{}

// Error implements the error interface
func (err APIError) Error() string {
	errs := []string{}
	for k, v := range err {
		errs = append(errs, fmt.Sprintf("%s: %v", k, v))
	}
	return strings.Join(errs, "; ")
}

// ErrRequestFailed creates an APIError from the
// HTTP response.
func ErrRequestFailed(res *http.Response) APIError {
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return APIError{"error": err.Error()}
	}
	apiErr := APIError{}
	if err := json.Unmarshal(body, &apiErr); err != nil {
		return APIError{"error": body}
	}
	return apiErr
}
