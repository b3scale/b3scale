package bbb

import (
	"fmt"
	"net/url"
)

// Params for the BBB API
type Params map[string]interface{}

// Encode query parameters
func (p Params) Encode() string {
	var q []string
	for k, v := range p {
		vStr := url.QueryEscape(fmt.Sprintf("%v", v))
		q = append(q, fmt.Sprintf("%s=%s", k, vStr))
	}
	return strings.Join(q, "&")
}
