package store

// BackendSettings hold per backend runtime configuration.
type BackendSettings struct {
	Tags []string `json:"tags"`
}

// Merge with a partial update. Nil fields are ignored.
// If a field was updated this will return true
func (s *BackendSettings) Merge(update *BackendSettings) bool {
	updated := false
	if update.Tags != nil {
		s.Tags = update.Tags
		updated = true
	}

	return updated
}

// FrontendSettings hold all well known settings for a
// frontend.
type FrontendSettings struct {
	RequiredTags        []string                     `json:"required_tags"`
	DefaultPresentation *DefaultPresentationSettings `json:"default_presentation"`
}

// Merge with a partial update. Fields that are nil
// will be ignored.
func (s *FrontendSettings) Merge(update *FrontendSettings) bool {
	updated := false
	if update.RequiredTags != nil {
		s.RequiredTags = update.RequiredTags
		updated = true
	}
	if update.DefaultPresentation != nil {
		s.DefaultPresentation = update.DefaultPresentation
		updated = true
	}
	return updated
}

// DefaultPresentationSettings configure a per frontend
// default configuration
type DefaultPresentationSettings struct {
	URL   string `json:"url,omitempty"`
	Force bool   `json:"force,omitempty"`
}
