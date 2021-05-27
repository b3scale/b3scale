

# b3scale REST API

## Authentication

We accept JWTs for authentication.
A public key or shared secret should be provided
through the environment.


## Resources and Scope

All API resources are prefixed with the api version.

### Scope

The privileges of a request are determined through the
`scope` claim of the JWT.

The following scopes are defined: `b3scale` and `b3scale:admin`.
The admin scope is required for accessing resources outside the
current user identified by the `sub` claim.


### Resources

 /api/v1/frontends

    GET   :: Retrieve a list of frontends
          SC b3scale.frontends:list

    POST  :: Register a new frontend
          SC b3scale.frontends:create

 /api/v1/frontends/<id>

    GET    :: Retrieve the frontend.
    PATCH  :: Update the frontend. Only fields  provided in the
              request will be updated. This applies for the
              nested `settings` object aswell.
    DELETE :: Remove the frontend.
 
 /api/v1/backends

    GET   :: Retrieve a list of backends
    POST  :: Register a new backend

 /api/v1/backends/<id>

    GET    :: Retrieve the backend.
    PATCH  :: Update the backend. Only fields provided in the request
              will be updated. This applies for the nested `settings`
              object aswell.
    DELETE :: Remove the backend.

 /api/v1/meetings

    GET    :: Retrieve a list of meetings known to the cluster
    DELETE :: Stop all meetings matching the filter or scope.

    Filters:  backend_id, frontend_id

 /api/v1/meetings/<id>

    GET    :: Get the meeting state from the cluster
    DELETE :: Force Stop a meeting


