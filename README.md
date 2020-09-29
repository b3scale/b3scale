# B3scale

A scalelite replacement.

## Options

The following environment variables can be configured:

 * `B3SCALE_FRONTENDS` A path to a file with frontends,
   Default: `etc/b3scale/frontends.conf`
  
 * `B3SCALE_BACKENDS` A path to a file with frontends
   Default: `etc/b3scale/backends.conf`
 
 * `B3SCALE_LISTEN_HTTP` Accept http connections here

 * `B3SCALE_REDIS_ADDRS` A space separated list of
   redis addresses. E.g. ":7001 :7002 :7003"
   Default: ":6379"

 * `B3SCALE_REDIS_USERNAME` A username for the redis cluster
 * `B3SCALE_REDIS_PASSWORD` A password for the redis cluster

