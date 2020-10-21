
--
-- Create the initial b3scale db schema
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE backends (
    id      uuid DEFAULT uuid_generate_v4() PRIMARY KEY,

    host    text NOT NULL,
    secret  text NOT NULL,

    tags    text ARRAY,

    -- Timestamps
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NULL DEFAULT NULL
);



