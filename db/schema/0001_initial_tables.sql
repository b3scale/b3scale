
--
-- Create the initial b3scale db schema
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Backends
CREATE TABLE backends (
    id      uuid DEFAULT uuid_generate_v4() PRIMARY KEY,

    host    text NOT NULL,
    secret  text NOT NULL,

    tags    text ARRAY,

    -- Timestamps
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NULL DEFAULT NULL
);

-- Frontends
CREATE TABLE frontends (
    id      uuid DEFAULT uuid_generate_v4() PRIMARY KEY,

    key     text NOT NULL UNIQUE,
    secret  text NOT NULL

    -- Timestamps
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NULL DEFAULT NULL
);


-- Meetings
CREATE TABLE meetings (
    id      uuid PRIMARY KEY,

    -- All meeting data is stored in the jsonb field.
    -- This should be sufficient for now.
    -- A later optimization could be moving out the attendees
    -- into a separate entity.
    info    jsonb NOT NULL,

    -- Relations
    frontend_id uuid NULL DEFAULT NULL
                REFERENCES frontends(id)
                ON DELETE SET NULL,

    backend_id uuid NOT NULL
               REFERENCES backends(id)
               ON DELETE CASCADE,

    -- Timestamps
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NULL DEFAULT NULL
);

-- Recordings

