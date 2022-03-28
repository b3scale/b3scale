
## Request Handling

Incoming requets are decoded by an echo middleware.
If a request matches a bbb api request, then a
bbb.Request is created and passed to the 
cluster.

  [HTTP API] -> [RequestDecoder] ->
      [Dispatch]  -> [ResponseEncode] -> [HTTP API]

There are now two middleware chains in Dispatch:
The router chain and the request chain.

The router is chain filters and removes potential backends.
The request chain modifies and dispatches the request
to a backend.

A middleware is a handler function, accepting the next
handler function as argument.
