package store

import "github.com/b3scale/b3scale/pkg/bbb"

// Tags are a list of strings with labels to declare
// for example backend capabilities
type Tags []string

// BackendSettings hold per backend runtime configuration.
type BackendSettings struct {
	Tags Tags `json:"tags,omitempty"`
}

// DefaultPresentationSettings configure a per frontend
// default presentation.
type DefaultPresentationSettings struct {
	URL   string `json:"url"`
	Force bool   `json:"force"`
}

// FrontendSettings hold all well known settings for a
// frontend.
type FrontendSettings struct {
	RequiredTags        Tags                         `json:"required_tags,omitempty"`
	DefaultPresentation *DefaultPresentationSettings `json:"default_presentation,omitempty"`

	CreateDefaultParams  bbb.Params `json:"create_default_params,omitempty" doc:"Provide key value params, which will be used as a default when a meeting is created. See the BBB api documentation for which params are valid. The param value must be encoded as string."`
	CreateOverrideParams bbb.Params `json:"create_override_params,omitempty" doc:"A key value set of params which will override parameters from the frontend when a meeting is created."`
}
