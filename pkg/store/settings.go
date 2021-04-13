package store

// Tags are a list of strings with labels to declare
// for example backend capabilities
type Tags []string

// BackendSettings hold per backend runtime configuration.
type BackendSettings struct {
	Tags Tags `json:"tags,omitempty"`
}

// FrontendSettings hold all well known settings for a
// frontend.
type FrontendSettings struct {
	RequiredTags        Tags                         `json:"required_tags,omitempty"`
	DefaultPresentation *DefaultPresentationSettings `json:"default_presentation,omitempty"`
}

// DefaultPresentationSettings configure a per frontend
// default presentation.
type DefaultPresentationSettings struct {
	URL   string `json:"url"`
	Force bool   `json:"force"`
}
