-- Runs once when the local Postgres volume is first created.
-- LOCAL DEVELOPMENT ONLY: the password is public in this repo.
-- The retention job deletes change history older than HISTORY_RETENTION_DAYS.
CREATE ROLE retention LOGIN PASSWORD 'retention';
