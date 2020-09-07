package cluster

// Schema is a mapping of variable names and decode hints
type Schema map[string]string

// RequestHandler accepts a bbb request and state. It produces
// a bbb response or an error.
type RequestHandler func(*Request) (Response, error)

// RequestMiddleware is a plain middleware without a state
type RequestMiddleware func(next RequestHandler) RequestHandler

// RouterHandler accepts a bbb request and returns
// a bbb request.
type RouterHandler func(*Request) (*Request, error)

// A RouterMiddleware accepts a handler function
// and returns a decorated handler function.
type RouterMiddleware func(next RouterHandler) RouterHandler
