
## Request Handling

Incoming requets are decoded by an echo middleware.
If a request matches a bbb api request, then a
bbb.Request is created and passed to the 
cluster.

  [HTTP API] -> [RequestDecoder] ->
      [Gateway]  -> [ResponseEncode] -> [HTTP API]

Let's have a look at the gateways's `Dispatch` function,
handling a decoded request:

Dispatch puts the request in a middleware.

The middleware chain can the modify the request and
the response.

A middleware is a handler function, accepting the next
handler function as argument.

      Middleware Stack:

    
    Request -> [user middlewares] -> ...

                                  -> [filter] -> [loader]
     ...      -> [dispatch/merge] -> [filter] -> [loader]
                                  -> [filter] -> [loader]



               


