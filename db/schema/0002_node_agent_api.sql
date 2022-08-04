

--
-- NodeAgent API
--

CREATE TABLE agents (
    id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,

    -- The agent is assigned to a backend with an
    -- individual api secret. The secret will be used to
    -- generate an individual access token for a node agent.
    
);

