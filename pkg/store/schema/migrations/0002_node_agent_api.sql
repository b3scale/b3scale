

--
-- NodeAgent API
--
-- %% Author: annika
-- %% Date: 2022-08-05
--

ALTER TABLE backends
  ADD agent_ref VARCHAR(64) NULL UNIQUE;
 
