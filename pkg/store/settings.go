package store

// Settings hold per front or backend runtime configuration.
// Variables can be accessed during request routing and
// handling in middlewares.
type Settings struct {
	Tags struct {
		Required []string `json:"required,omitempty"`
		Provided []string `json:"provided,omitempty"`
	} `json:"tags,omitempty"`
}
