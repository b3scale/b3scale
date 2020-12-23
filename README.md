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

    $ b3scalectl set backend --state ready https://bbbb01.example.net/bigbluebutton/api/

The host should match the one you see with

    $ b3scalectl show backends


## Deleting Backends

Backends can be removed through

    $ b3scalectl rm backend https://bbbb01.example.net/bigbluebutton/api/
    
This will initiate a decomissioning process, where the backend will not longer
be used for creating new sessions.

It will be permanently deleted after the last session was closed.


### Issues

At the moment a running node agent might prevent the backend from being deleted.


