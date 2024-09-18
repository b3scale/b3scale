package callbacks

// Callbacks are used to notify the client.
// I strongly dislike this. Better solutions would have
// been: polling an endpoint or server sent events for
// subscriptions.
//
// Callbacks are always prone to inconsistencies, or
// race conditions, or can get lost.

import (
	"fmt"
	"net/url"
)

// A Callback contains a singl attribute
// in the body: signed_parameters.
type Callback struct {
	SignedParameters string `json:"signed_parameters"`
}

// Validate checks an OnRecordingReady request
func (c *Callback) Validate() error {
	if c.SignedParameters == "" {
		return fmt.Errorf("signed_parameters is required")
	}
	return nil
}

// Encode encodes the callback request as
// Form-Encoded data.
func (c *Callback) Encode() string {
	values := url.Values{}
	values.Add("signed_parameters", c.SignedParameters)
	return values.Encode()
}
