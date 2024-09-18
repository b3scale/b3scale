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

// SignedBody contains signed parameters posted
// by the bbb node agent to the callback URL.
//
// The signed_parameters attribute is a JWT.
type SignedBody struct {
	SignedParameters string `json:"signed_parameters" form:"signed_parameters"`
}

// Validate checks an OnRecordingReady request
func (b *SignedBody) Validate() error {
	if b.SignedParameters == "" {
		return fmt.Errorf("signed_parameters is required")
	}
	return nil
}

// Encode encodes the callback request as
// Form-Encoded data.
func (b *SignedBody) Encode() string {
	values := url.Values{}
	values.Add("signed_parameters", b.SignedParameters)
	return values.Encode()
}
