package store

// BackendSettings hold per backend runtime configuration.
type BackendSettings struct {
	Tags []string `json:"tags"`
}

// FrontendSettings hold all well known settings for a
// frontend.
type FrontendSettings struct {
	RequiredTags        []string `json:"required_tags"`
	DefaultPresentation struct {
		URL   string `json:"url,omitempty"`
		Force bool   `json:"force,omitempty"`
	}
}
