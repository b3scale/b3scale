package middlewares

import (
	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// A StateUpdate is a mapping of string
// keys and values.
type StateUpdate map[string]interface{}

// The State of the middleware
type State interface {
	// Apply a state update.
	Update(next StateUpdate) error

	// Schema creates a mapping of acceptable
	// keys and datatypes in an update.
	Schema() map[string]string

	// ID indentifies the state, so it
	// can be queried from an API
	ID() string
}

// HandlerFunc accepts a bbb request and state. It produces
// a bbb response or an error.
type HandlerFunc func(*bbb.Request, State) (*bbb.Response, error)

// A Middleware is a function, accepting a next handler function
// and returning a handler function and an initial state.
type Middleware func(next HandlerFunc) (HandlerFunc, State)
