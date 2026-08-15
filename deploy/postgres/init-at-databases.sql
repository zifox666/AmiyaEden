-- Executed by the official PostgreSQL image only when the data directory is
-- first initialized. The psql \gexec form makes both CREATE DATABASE actions
-- idempotent when this file is run manually during an existing installation.
SELECT 'CREATE DATABASE at_data'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'at_data')\gexec

SELECT 'CREATE DATABASE amiya_eden_sde'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'amiya_eden_sde')\gexec
