
# b3scale database schema

We use postgres to store a shared state across instances.

## Initialization

To initialize an empty database, you can use the `init.sh`
script to apply all sql scripts in `./schema/`

## Tables

### backends

The backends table holds a list of backend configurations
in the following attributes:
    
    id: backend identifier {uuid} (string)
    host: fqdn and api path of the bbb instance (string)
    secret: the backends secret in plain text
    tags: list of tags ([]string)

### frontends

This table holds all frontend configurations:

    id: identifier {uuid} (string)
    key: name (e.g. bigbluebutton) (string)
    secret: the frontend secret in plain text


### store

Holds meetings and recordings data or other shared
backend state. Has key value store properties.

### commands (...)

The commands table acts as a queue for actions performed
by the instance. e.g. reloading the configuration.

    id: uuid of the job (string)
    sequence: e.g. timestap nano (uint64)
    action: name of the action (string)
    args: list of arguments ([json])


