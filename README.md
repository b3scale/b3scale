# B3scale

[![Test](https://github.com/b3scale/b3scale/actions/workflows/main.yml/badge.svg)](https://github.com/b3scale/b3scale/actions/workflows/main.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/b3scale/b3scale)](https://goreportcard.com/report/github.com/b3scale/b3scale)


An efficient multi tenant load balancer for BigBlueButton.

## About

`b3scale` is a load balancer designed to be used in place of scalelite. Work was
started in 2020 to provide multiple features not possible before:

  * *API driven*: REST API allows integrating b3scale straight to your CRM
    and/or user portal
  * *Observable*: Prometheus endpoint for all essential operational information
  * *Multi tenancy*: b3scale introduces the concept of frontends
    * custom client secret
    * optional custom presentation
  * *Efficient*: 5 figure users with a single instance
  * *High availability and easy scale out*: Just add more b3scale servers and
    use with your existing HA solution
  * *Flexible backend handling*:
    * Map frontends to BBB nodes (backends)
    * Retire backends for updates
    * Powerful tagging system allows for friendly user testing, experiments and
      assignments depending on expected customer load
  * *True load-awareness*: using reports from the `b3scalenoded` agent, b3scale
    can schedule meetings more efficiently
  * *Easy deployment*: b3scale is written in Go: no dependencies, just deploy a
    single binary

## Basic principle

To discuss the principal design of b3scale, consider the following schematic:

![b3scale architecture](doc/b3scale-architecture.png)

b3scale services different *frontends*. Those can be standard apps such as
Greenlight, Nextcloud or Moodle, but can also really be anything that implements
the BigBlueButton API, even custom web apps.

A frontend can initiate a new meeting via b3scale, which will assign it to a
*backend* node and will keep track of the assignment. Users joining will thus
be assigned to the correct backend.

Using *tags* you can assign specific roles to one or more backend nodes: you
can assign a customer to a specific set of nodes, effectively forming a
dedicated cluster. It is also possible to  steer friendly users towards nodes
that contain experimental features.

You can take a backend offline by disabling it. This will not affect currently
running meetings. It will only remove the node from consideration for new
meetings. This way, backend nodes can be drained e.g. in preparation for
scheduled maintenance.

For backends nodes, b3scale provides `b3scalenoded`, an agent for backend nodes
that monitors certain parameters straight from redis and reports them to
b3scale in an inexpensive, resource conserving fashion.

*DEPRECATION NOTICE:*
The `b3scalenoded` will be deprecated in favour of the `b3scaleagent`,
which does the same thing, but uses the HTTP API.
 
Please note, that the agents need unique access tokens for each
backend.

A new access token can be crated using `b3scalectl auth authorize_node_agent`.


## Terminology

* *Frontend:* A BigBlueButton frontend such as Greenlight
* *Backend:* A BigBlueButton server to distribute meetings on
* *Middleware:* Different middleware implementations provide different aspects
  of b3scale's core functionality such as tagging

## API documentation

Please find the API documentation in the [REST API](doc/rest_api.md) file.

## Configuration

b3scale daemons are configured through environment variables and do not use a config file. Example environment files for use with Docker, Kubernetes or systemd with all eligable settings can be found here:

* [Environment for b3scaled](doc/example.env.b3scaled)
* [Environment for b3scaleagent](doc/example.env.b3scaleagent)
* [Environment for b3scalenoded](doc/example.env.b3scalenoded)

Find more documentation below.

### Environment variables

 * `B3SCALE_LISTEN_HTTP` Accept http connections here.
    Default: `127.0.0.1:42353`

 * `B3SCALE_DB_URL` is the connect string passed to the
    database connection.  Default is `postgres://postgres:postgres@localhost:5432/b3scale`

    You can use either the DSN format or an URL format:

    ### Example DSN
    `user=jack password=secret host=pg.example.com port=5432 dbname=mydb sslmode=verify-ca`

    ### Example URL
    `postgres://jack:secret@pg.example.com:5432/mydb?sslmode=verify-ca`

    (b3scaled and b3scalenoded only)

 * `B3SCALE_DB_POOL_SIZE` the number of maximum parallel connections
    we will allocate. Please note that one connection per request will
    be blocked and returned to the pool afterwards.

    Default: 128

 * `B3SCALE_LOG_LEVEL` set the log level. Possible values are:

        panic  5
        fatal  4
        error  3
        warn   2
        info   1
        debug  0
        trace -1

    You can use either the numeric or integer value

  * `B3SCALE_LOG_FORMAT` choose between `plain` or `structured` logging.
     The default is `structured` and will emit JSON on stderr.

Same applies for the `b3scalenoded`, however only `B3SCALE_DB_URL`
is required.

The `b3scalenoded` and `b3scaleagent` read from the same configuration as BigBlueButton, the environment variable for the file is:

 * `BBB_CONFIG`, which defaults to:
    `/usr/share/bbb-web/WEB-INF/classes/bigbluebutton.properties`

This file must be readable for the b3scalenoded.

 * `B3SCALE_ACCESS_TOKEN` stores the authorized access token for a
    node.

You can authorize a new agent using b3scalectl:

    b3scalectl auth authorize_node_agent

Unless the `B3SCALE_API_JWT_SECRET` environment variable is set,
you will be prompted to enter the API secret.

As an alternative the secret can be provided through the `--secret`
flag when authorizing a new agent.

A custom agent identifier can be provided through the `--ref` option.
If none is provided, it will be generated.

A full example would be

    b3scalectl auth authorize_node_agent --ref backend23 --secret my-api-secret

Please note, that the `ref` identifier must be unique, as only one backend
can be associated with an agent.

 * `B3SCALE_API_URL` must be provided for the `b3scaleagent` to find the API.
    Only the host part is required: e.g. `https://b3scale.example/`

The load factor of the backend can be set through:

 * `B3SCALE_LOAD_FACTOR` (default `1.0`)

 * `B3SCALE_API_JWT_SECRET` if not empty, the API will be enabled
    and accessible through /api/v1/... with a JWT bearer token.
    You can set the jwt claim `scope` to `b3scale:admin` to create
    an admin token. You can generate an access token using `b3scalectl`:

       b3scalectl auth create_access_token --sub node42 --scopes b3scale,b3scale:admin,b3scale:node

    You are then prompted to paste the `B3SCALE_API_JWT_SECRET`.

    You can pass the secret through the --secret longopt - however this is discouraged
    because it might end up in the shell history. Be careful.

    In case your `b3scalectl` responds with

        3:43PM FTL this is fatal error="message: invalid or expired jwt"

    remove the access token in `~/.config/b3scale/<host>.access_token`

 
 * `B3SCALE_RECORDINGS_PUBLISHED_PATH` required if recordings are supported: This points to
   the shared path where published recordings are.
   Example: `/ceph/recordings/published`

 * `B3SCALE_RECORDINGS_UNPUBLISHED_PATH` recordings are moved here, when unpublished
   Please note that in both directories the subfolder for the format should
   be present. (e.g. `/ceph/recordings/unpublished/presentation`)
   Example: `/ceph/recordings/unpublished`

 * `B3SCALE_RECORDINGS_PLAYBACK_HOST` path to host with the player.
   For example: https://playback.mycluster.example.bbb/

## Adding Backends

### Using the node agent

Adding a backend using the node agent `b3scalenoded` / `b3scaleagent`
can be as simple as starting it with the `-register` option for autoregistering the
new node.

For the `b3scaleagent` the `B3SCALE_ACCESS_TOKEN` and `B3SCALE_API_URL` needs to be
provided.

The node is identified through the `BBB_CONFIG` file from bbb-web.

Autoregistering is turned off by default.

After registering the node, you have to enable it.

The default `admin_state` of the node is init. To enable the
node, set the admin state to `ready`.

    $ b3scalectl enable backend https://bbbb01.example.net/bigbluebutton/api/

The host should match the one you see with

    $ b3scalectl show backends


## Disable Backends

You can exclude backends as targets for new meetings
by running

    $ b3scalectl disable backend https://bbbb01.example.net/bigbluebutton/api/


## Deleting Backends

Backends can be removed through

    $ b3scalectl rm backend https://bbbb01.example.net/bigbluebutton/api/

This will initiate a decomissioning process, where the backend will not longer
be used for creating new sessions.

It will be permanently deleted after the last session was closed.

## Disable/Enable frontends

Frontends can be disabled without removing it completly with

`$ b3scalectl disable frontend frontend1`

Disabled frontends can be enabled again with

`$ b3scalectl enable frontend frontend1`


## Middleware Configuration

The middlewares can be configured using b3scalectl or via API calls.
A property value will be interpreted as JSON.

### Configure tagged routing


    b3scalectl set backend -j '{"tags": ["asdf", "foo", "bar"]}' https://backend23/
    b3scalectl set frontend -j '{"required_tags": ["asdf"]}' frontend1

Unset a value with explicit null:

    b3scalectl set frontend -j '{"required_tags": null}' frontend1

### Configure a default presentation

    b3scalectl set frontend -j '{"default_presentation": {"url": "https://..."}}' frontend1

### Configure create parameter *defaults* and *overrides*

An *override* will replace the parameter of the request.

A *default* is added to the request parameters if not present.
In case of `disabledFeatures`, the list coming from the request
will be amended with the defaults. If no disabledFeatures are
provided from the frontend, the defaults will be used.

All parameter values are strings and need to be encoded
according to the specifications in 
https://docs.bigbluebutton.org/dev/api.html#create

Some examples:

Set a default logo (if not present) and force some disabled features.
Addional disabled features from the frontend will be preserved.

    b3scalectl set frontend -j '{"create_default_params": {"logo": "logoURL", "disabledFeatures": "chat,captions"}}' frontend1


Force disable recordings:

    b3scalectl set frontend -j '{"create_override_params": {"allowStartStopRecording": "false", "autoStartRecording": "false"}}' frontend1


Set disabledFeatures, ignoring requested disabled features from the frontend:

    b3scalectl set frontend -j '{"create_override_params": {"disabledFeatures": "captions"}}' frontend1


Setting `create_default_params` or `create_override_params` is always
a replacement of the current value. If `null` is provided, the configuration
key will be unset.

    b3scalectl set frontend -j '{"create_override_params": null, "create_default_params": null}' frontend1

## Monitoring

Metrics are exported in a `prometheus` compatible format under `/metrics`.

## Bug reports and Contributions

If you discover a problem with b3scale or have a feature request, please open a
[bug report](https://github.com/b3scale/b3scale/issues/new). Please
check the [existing
issues](https://github.com/b3scale/b3scale/issues) before reporting
new ones. Do not start work on new features without prior discussion. This
helps us to coordinate development efforts. Once your feature is discussed,
please file a merge request for the `develop` branch. Merge requests to
`main`happen from `develop` only.

## Disclaimer

*This project uses BigBlueButton and is not endorsed or certified by
BigBlueButton Inc. BigBlueButton and the BigBlueButton Logo are trademarks of
BigBlueButton Inc.*
