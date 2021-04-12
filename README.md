# B3scale

A scalelite replacement.

## Options

The following environment variables can be configured:

 * `B3SCALE_LISTEN_HTTP` Accept http connections here.
    Default: `127.0.0.1:42353`

 * `B3SCALE_DB_URL` is the connect string passed to the
    database connection.  Default is `postgres://postgres:postgres@localhost:5432/b3scale`

    You can use either the DSN format or an URL format:

    ### Example DSN
    `user=jack password=secret host=pg.example.com port=5432 dbname=mydb sslmode=verify-ca`

    ### Example URL
    `postgres://jack:secret@pg.example.com:5432/mydb?sslmode=verify-ca`


 * `B3SCALE_DB_POOL_SIZE` the number of maximum parallel connections
    we will allocate. Please note that one connection per request will
    be blocked and returned to the pool afterwards.

    Default: 128

 * `B3SCALE_REVERSE_PROXY_MODE` if set to `yes` or `1` or `true` it will
   enable the reverse proxy mode in the cluster gateway.
   You have to configure a reverse proxy e.g. nginx to handle
   subsequent client requests.

 * `B3SCALE_LOG_LEVEL` set the log level. Possible values are:

        panic  5
        fatal  4
        error  3
        warn   2
        info   1
        debug  0
        trace -1
    
    You can use either the numeric or integer value


Same applies for the `b3scalenoded`, however only `B3SCALE_DB_URL`
is required.

The `b3scalenoded` uses the same configuration as BigBlueButton,
the environment variable for the file is:

 * `BBB_CONFIG`, which defaults to:
    `/usr/share/bbb-web/WEB-INF/classes/bigbluebutton.properties`
    
This file must be readable for the b3scalenoded.

The load factor of the backend can be set through:

 * `B3SCALE_LOAD_FACTOR` (default `1.0`)


## Adding Backends

### Using the node agent

Adding a backend using the node agent `b3scalenoded` can be as simple
as starting it with the `-register` option for autoregistering the
new node.

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


## Middleware Configuration

The middlewares can be configured using b3scalectl:
A property value will be interpreted as JSON or string.

    b3scalectl set frontend --prop 'my_middleware.subkey={"a": [23, 42]}' $FRONTEND

    b3scalectl set frontend --prop 'my_middleware=test123' $FRONTEND

Unset / delete a property:

    b3scalectl set frontend --prop 'my_middleware=' $FRONTEND



### Issues

At the moment a running node agent might prevent the backend from being deleted.


