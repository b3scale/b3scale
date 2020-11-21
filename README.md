# B3scale

A scalelite replacement.

## Options

The following environment variables can be configured:

 * `B3SCALE_LISTEN_HTTP` Accept http connections here
    Default: 127.0.0.1:42353

 * `B3SCALE_LISTEN_HTTP2` Accept http2 connections here
    Default: 127.0.0.1:42352

 * `B3SCALE_DATABASE_URL` is the connect string passed to the
    database connection.  Default is `postgres://postgres:postgres@localhost:5432/b3scale`

    You can use either the DSN format or an URL format:

    # Example DSN
    user=jack password=secret host=pg.example.com port=5432 dbname=mydb sslmode=verify-ca

    # Example URL
    postgres://jack:secret@pg.example.com:5432/mydb?sslmode=verify-ca



