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
    -- accept any requests.
    'stopped',
    -- The backend is disabled and should not accept
    -- any requests.
    'decommissioned'
    -- The backend is marked for removal and should
    -- not accept new requests.
);


-- Backends
CREATE TABLE backends (
    id      uuid DEFAULT uuid_generate_v4() PRIMARY KEY,

    -- The instance state indicates the current state
    -- of the bbb node, the admin state declares the
    -- desired state of the instance.
    node_state  instance_state NOT NULL DEFAULT 'init',
    admin_state instance_state NOT NULL DEFAULT 'stopped',

    -- Heartbeat of the node agent: consider the node dead
    -- if this is older than a given threashold.
    agent_heartbeat TIMESTAMP NOT NULL DEFAULT '2000-01-01 00:00:00',

    -- Statistics: We make routing decision based on
    -- these numbers
    latency         BIGINT   NOT NULL DEFAULT 0.0,
    meetings_count  INTEGER  NOT NULL DEFAULT 0,
    attendees_count INTEGER  NOT NULL DEFAULT 0, 

    load_factor     REAL     NOT NULL DEFAULT 1.0,

    last_error text NULL DEFAULT NULL,

    host    text NOT NULL      UNIQUE,
    secret  text NOT NULL,

    -- Runtime configuration can be added for each backend
    -- and can then be accessed from within middlewares
    settings  jsonb NOT NULL DEFAULT '{}'::jsonb,

    -- Timestamps
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    synced_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP 
);


-- Frontends
CREATE TABLE frontends (
    id      uuid DEFAULT uuid_generate_v4() PRIMARY KEY,

    key     text NOT NULL UNIQUE,
    secret  text NOT NULL,

    -- The account reference will be used to limit
    -- access to the list of frontends when accessed
    -- through an authorized API request.
    account_ref VARCHAR(80) NULL DEFAULT NULL,

    -- Runtime configuration can be added for each
    -- frontend. This should be used to save middleware
    -- settings like for example overriding a default
    -- presentation with a custom slide on create.
    settings  jsonb NOT NULL DEFAULT '{}'::jsonb,

    active    BOOLEAN NOT NULL DEFAULT true,

    -- Timestamps
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    synced_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP 
);

CREATE INDEX idx_frontends_account_ref ON frontends
 USING HASH ( account_ref );
  

-- Frontend Meetings:
-- We need to keep track of meeting IDs associated with
-- a frontend for associating recordings with the proper
-- frontend even after the meeting (state) is already gone
-- and not longer present on a backend.
CREATE TABLE frontend_meetings (
    frontend_id uuid         NOT NULL
                REFERENCES   frontends(id)
                ON DELETE    CASCADE,

    meeting_id  VARCHAR(255) NOT NULL UNIQUE,

    -- For housekeeping we should track if this
    -- is recent data or can maybe be deleted.
    seen_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


-- Meetings:
-- The store tables: `meetings`, `recordings`, the shared state
-- between instances. 
--
-- Please note that the primary source of truth
-- about meetings etc. should be the bbb instance.
--
-- Also: If required this could be split up
-- into dedicated entities.
--
CREATE TABLE meetings (
    -- The BBB meeting ID, and internal ID
    id          VARCHAR(255) PRIMARY KEY,
    internal_id VARCHAR(255) UNIQUE,

    -- All state data is stored in the jsonb field.
    -- This should be sufficient for now; if required
    -- the state could be broken up into meetings,
    -- attendees, recordings, etc...
    state   jsonb NOT NULL,

    -- Relations
    frontend_id uuid       NULL
                REFERENCES frontends(id)
                ON DELETE  CASCADE,

    backend_id uuid        NULL
               REFERENCES  backends(id)
               ON DELETE   SET NULL,

    -- Timestamps
    created_at  TIMESTAMP  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    synced_at   TIMESTAMP  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Recordings are quite like meetings, however
-- a foreign key relation exists to improve querying.
CREATE TABLE recordings (
    -- The BBB record ID
    state   jsonb   NOT NULL,

    text_track_states jsonb NOT NULL DEFAULT '[]',

    record_id  VARCHAR(255) NOT NULL PRIMARY KEY,

    -- Relations
    frontend_id uuid NOT NULL
                REFERENCES frontends(id)
                ON DELETE CASCADE,

    meeting_id          VARCHAR(255) NOT NULL,
    internal_meeting_id VARCHAR(255) NOT NULL,

    -- Timestamps
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    synced_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_recordings_meeting_id ON recordings (meeting_id);
CREATE INDEX idx_recordings_internal_meeting_id ON recordings (internal_meeting_id);


-- Commands state transition between requested
-- and a final success indicator.
CREATE TYPE command_state AS ENUM (
    -- The initial state of a command is that the request
    -- was issued.
    'requested',
    -- The command is being executed
    'running',
    -- Finally executing the command was successful
    'success',
    -- Finally executing the command was not successful
    'error'
);

-- Commands are jobs processed by any b3scale instance.
-- B3scale instances listen on a pg notify queue.
CREATE TABLE commands (
    id      uuid          DEFAULT uuid_generate_v4()
                          PRIMARY KEY,
    seq     SERIAL,

    state   command_state DEFAULT 'requested',

    -- The action encodes a invokable function
    -- for example retrieving the current bbb state from
    -- a backend.
    action  VARCHAR(80)   NOT NULL,
    params  json          NULL,
    result  json          NULL,

    -- Job control: A deadline is required for each 
    -- command. Afterwards the command is expired.
    deadline   TIMESTAMP  NOT NULL,
    started_at TIMESTAMP  NULL       DEFAULT NULL,
    stopped_at TIMESTAMP  NULL       DEFAULT NULL,
    created_at TIMESTAMP  NOT NULL   DEFAULT CURRENT_TIMESTAMP
);

-- AfterCommandsInsert
-- Procedure to be called for every new command
CREATE FUNCTION after_commands_insert() RETURNS TRIGGER AS $$
BEGIN
  -- Housekeeping: Remove expired commands.
  DELETE FROM commands
   WHERE (deadline + interval '1 minute') 
         < now() AT TIME ZONE 'utc';

  -- Finally inform instances, that a new command
  -- was queued.
  NOTIFY commands_queue;
  RETURN NULL;
END
$$ LANGUAGE plpgsql;

CREATE TRIGGER  command_insert    AFTER INSERT ON commands
  FOR EACH ROW  EXECUTE PROCEDURE after_commands_insert();

-- The meta table stores information about the schema
-- like when it was migrated and the current revision.
CREATE TABLE __meta__ (
    version     INTEGER   NOT NULL  UNIQUE,
    description TEXT      NOT NULL,
    applied_at  TIMESTAMP NOT NULL  DEFAULT CURRENT_TIMESTAMP
);

