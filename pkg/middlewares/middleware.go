package middlewares

import (
	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// HandlerFunc accepts a bbb request and creates a bbb response
type HandlerFunc func(*bbb.Request) (*bbb.Response, error)

// A Middleware is a function, accepting a next handler function
// and returning a handler function
type Middleware func(next HandlerFunc) HandlerFunc
