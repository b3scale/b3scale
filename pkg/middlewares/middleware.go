package middlewares

import (
	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// A StateUpdate is a mapping of string
// keys and values.
type StateUpdate map[string]interface{}

// The State of the middleware
type State interface {
	Update(next StateUpdate) error
	Schema() map[string]string
}

// HandlerFunc accepts a bbb request and creates a bbb response
type HandlerFunc func(*bbb.Request) (*bbb.Response, error)

// A Middleware is a function, accepting a next handler function
// and returning a handler function and an initial state.
type Middleware func(next HandlerFunc) (HandlerFunc, State)
