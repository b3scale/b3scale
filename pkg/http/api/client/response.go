package client

import (
	"encoding/json"
	"io"
	"net/http"
)

// JSON helper
func readJSONResponse(res *http.Response, data interface{}) error {
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, data)
}

// Status helper
func httpSuccess(res *http.Response) bool {
	s := res.StatusCode
	return s >= 200 && s < 400
}
