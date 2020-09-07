package cluster

import (
	"fmt"
)

// Schema is a mapping of variable names and decode hints
type Schema map[string]string

// The Handler state interface
type Handler interface {
	// Update the handler state
	Update(update interface{}) error

	// Schema creates a mapping of acceptable
	// keys and datatypes in an update.
	Schema() Schema

	// ID indentifies the handler, so it
	// can be queried from an API
	ID() string

	// A Middleware is a function, accepting a next
	// handler function and returning a handler function
	Middleware(next HandlerFunc) HandlerFunc
}

// HandlerFunc accepts a bbb request and state. It produces
// a bbb response or an error.
type HandlerFunc func(*Request) (Response, error)

// MiddlewareFunc is a plain middleware without a state
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

// nilHandler is an empty handler, that only will result
// in an error when called.
func nilHandler(_req *Request) (Response, error) {
	return nil, fmt.Errorf("end of middleware chain")
}
