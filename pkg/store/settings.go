package store

import "github.com/b3scale/b3scale/pkg/bbb"

// Tags are a list of strings with labels to declare
// for example backend capabilities
type Tags []string

// BackendSettings hold per backend runtime configuration.
type BackendSettings struct {
	Tags Tags `json:"tags,omitempty" doc:"The backend provides these tags. A frontend can require a list of tags. This can be used to dedicate parts of the cluster."`
}

// DefaultPresentationSettings configure a per frontend
// default presentation.
type DefaultPresentationSettings struct {
	URL   string `json:"url" doc:"An URL pointing to a default presentation." example:"https://assets.mycluster.example.com/tenant1235/presentation.pdf"`
	Force bool   `json:"force" doc:"Override any default presentation provided by the frontend."`
}
type DefaultLimitAttendeesSettings struct {
  Limit   int `json:"limit" doc:"Limit of overall attendees for a frontend."`
}

// FrontendSettings hold all well known settings for a
// frontend.
type FrontendSettings struct {
	RequiredTags        Tags                         `json:"required_tags,omitempty" doc:"When selecting a backend for creating a meeting, only consider nodes providing all of the required tags."`
	DefaultPresentation *DefaultPresentationSettings `json:"default_presentation,omitempty"`
	LimitAttendees *DefaultLimitAttendeesSettings `json:"limit_attendees,omitempty"`

	CreateDefaultParams  bbb.Params `json:"create_default_params,omitempty" doc:"Provide key value params, which will be used as a default when a meeting is created. See the BBB api documentation for which params are valid. The param value must be encoded as string."`
	CreateOverrideParams bbb.Params `json:"create_override_params,omitempty" doc:"A key value set of params which will override parameters from the frontend when a meeting is created."`
}
