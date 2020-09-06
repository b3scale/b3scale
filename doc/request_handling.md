
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

    Request -> [<user middlewares>] -> ...

                                 -> [filter] -> [apiBackend]
     ...      -> [dispatchMerge] -> [filter] -> [apiBackend]
                                 -> [filter] -> [apiBackend]

A Handler is a stateful middleware.
It should be initialized with a constructor and should
safe and restore it's operational state if required.

The Handler interface exposes functions for retrieving
a schema for mapping updateable variables and types...

Handlers can be registered at a HttpController where
the current state(s) can be queried and
changes can be send to the endpoint...

API (draft):

 GET /api/v1/handlers
     /api/v1/handlers/<myHandler>
     
 OPTIONS /api/v1/handlers/<myHandler>
     Returns the update schema
 
 PUT /api/v1/handlers/<myHandler>
     Accepts an update according to schema

Maybe. 







