--
-- ----------------------
-- b3scale schema v.1.0.0
-- ----------------------
--
-- %% Author:      annika
-- %% Description: Create the initial b3scale db schema.
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";


-- The instance state indicates if the backend
-- is ready to accept requests.
CREATE TYPE instance_state AS ENUM (
    'init',    
    -- The backend is initializing.
    'ready',   
    -- Ready for accepting requests.
    'error', 
    -- An error occured and we should not longer
    -- accept new requests.
    'stopped'
    -- The backend is disabled and should not accept
    -- any requests.
);


-- Backends
CREATE TABLE backends (
    id      uuid DEFAULT uuid_generate_v4() PRIMARY KEY,

    -- The instance state indicates the current state
    -- of the bbb node, the admin state declares the
    -- desired state of the instance.
    node_state  instance_state NOT NULL DEFAULT 'init',
    admin_state instance_state NOT NULL DEFAULT 'ready',

    host    text NOT NULL,
    secret  text NOT NULL,

    tags    text ARRAY,

    -- Timestamps
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NULL DEFAULT NULL,
    synced_at   TIMESTAMP NULL DEFAULT NULL
);

-- Frontends
CREATE TABLE frontends (
    id      uuid DEFAULT uuid_generate_v4() PRIMARY KEY,

    key     text NOT NULL UNIQUE,
    secret  text NOT NULL,

    active  BOOLEAN NOT NULL DEFAULT true,

    -- Timestamps
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NULL DEFAULT NULL
);


-- The store table holds the shared state
-- between instances. This can be meeting data
-- but also recording information.
--
-- Please note that the primary source of truth
-- about meetings etc. should be the bbb instance.
--
-- Also: If required this could be split up
-- into dedicated entities.
--
-- To for example get all meetings for a backend:
-- SELECT * FROM store 
--          WHERE backend_id = $1
--            AND key LIKE 'meeting:%'
--
CREATE TABLE store (
    id      uuid PRIMARY KEY,

    -- The stores have an additional key, which can
    -- be used for querying - eg.: meeting:2839102938012
    key     text NOT NULL,

    -- All state data is stored in the jsonb field.
    -- This should be sufficient for now; if required
    -- the state could be broken up into meetings,
    -- attendees, recordings, etc...
    state   jsonb NOT NULL,

    -- Relations
    frontend_id uuid NULL DEFAULT NULL
                REFERENCES frontends(id)
                ON DELETE SET NULL,

    backend_id uuid NOT NULL
               REFERENCES backends(id)
               ON DELETE CASCADE,

    -- Timestamps
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NULL DEFAULT NULL,
    synced_at   TIMESTAMP NULL DEFAULT NULL
);

CREATE INDEX idx_store_key ON store USING btree (key);
