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

    last_error text NULL DEFAULT NULL,

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


-- The store tables: `meetings`, `recordings`,
-- `recording_text_tracks` hold the shared state
-- between instances. 
--
-- Please note that the primary source of truth
-- about meetings etc. should be the bbb instance.
--
-- Also: If required this could be split up
-- into dedicated entities.
--
CREATE TABLE meetings (
    -- The BBB meeting ID
    id      uuid PRIMARY KEY,

    -- All state data is stored in the jsonb field.
    -- This should be sufficient for now; if required
    -- the state could be broken up into meetings,
    -- attendees, recordings, etc...
    state   jsonb NOT NULL,

    -- Relations
    frontend_id uuid NOT NULL
                REFERENCES frontends(id)
                ON DELETE CASCADE,

    backend_id uuid NOT NULL
               REFERENCES backends(id)
               ON DELETE CASCADE,

    -- Timestamps
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NULL DEFAULT NULL,
    synced_at   TIMESTAMP NULL DEFAULT NULL
);

-- Recordings are quite like meetings, however
-- a foreign key relation exists to improve querying.
CREATE TABLE recordings (
    -- The BBB record ID
    id      uuid PRIMARY KEY,
    state   jsonb NOT NULL,

    -- Relations
    backend_id uuid NOT NULL
               REFERENCES backends(id)
               ON DELETE CASCADE,
    
    meeting_id uuid NOT NULL
               REFERENCES meetings(id)
               ON DELETE CASCADE,

    -- Timestamps
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NULL DEFAULT NULL,
    synced_at   TIMESTAMP NULL DEFAULT NULL
);

-- RecordingTextTracks are associated with recordings
-- meetings through a foreign key relation for querying.
CREATE TABLE recording_text_tracks (
    -- The BBB record ID
    id      uuid PRIMARY KEY,
    state   jsonb NOT NULL,

    -- Relations
    record_id   uuid NOT NULL
                REFERENCES recordings(id)
                ON DELETE CASCADE,

    -- Timestamps
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NULL DEFAULT NULL,
    synced_at   TIMESTAMP NULL DEFAULT NULL
);


-- The eta table stores information about the schema
-- like when it was migrated and the current revision.
CREATE TABLE __meta__ (
    version     INTEGER   NOT NULL  UNIQUE,
    description TEXT      NOT NULL,
    applied_at  TIMESTAMP NOT NULL  DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO __meta__ (version, description)
     VALUES (1, 'initial tables');

