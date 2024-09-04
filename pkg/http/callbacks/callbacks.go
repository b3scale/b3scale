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

// Callback interface provides a Validate and
// an Encode method.
type Callback interface {
	Validate() error
	Encode() string
}

// OnRecordingReady is a callback request
// with a `signed_parameters` payload.
type OnRecordingReady struct {
	SignedParameters string `json:"signed_parameters" form:"signed_parameters"`
}

// Validate checks an OnRecordingReady request
func (r *OnRecordingReady) Validate() error {
	if r.SignedParameters == "" {
		return fmt.Errorf("signed_parameters is required")
	}
	return nil
}

// Encode encodes the callback request as
// Form-Encoded data.
func (r *OnRecordingReady) Encode() string {
	values := url.Values{}
	values.Add("signed_parameters", r.SignedParameters)
	return values.Encode()
}
