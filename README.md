# B3scale

A scalelite replacement.

## Options

The following environment variables can be configured:

 * `B3SCALE_LISTEN_HTTP` Accept http connections here.
    Default: `127.0.0.1:42353`
 * `B3SCALE_LISTEN_HTTP2` Accept http2 connections here.
    Default: `127.0.0.1:42352`

 * `B3SCALE_DB_URL` is the connect string passed to the
    database connection.  Default is `postgres://postgres:postgres@localhost:5432/b3scale`

    You can use either the DSN format or an URL format:

    ### Example DSN
    `user=jack password=secret host=pg.example.com port=5432 dbname=mydb sslmode=verify-ca`

    ### Example URL
    `postgres://jack:secret@pg.example.com:5432/mydb?sslmode=verify-ca`


Same applies for the b3scalenoded, however only `B3SCALE_DB_URL`
and a `BBB_REDIS_URL` is required.

  * `BBB_REDIS_URL` accepts an URL to the BBB redis server.
    Default: `redis://localhost:6379/1`

    ### URL schema
    `redis://<user>:<password>@<host>:<port>/<db_number>`

    `unix://<user>:<password>@</path/to/redis.sock>?db=<db_number>`





