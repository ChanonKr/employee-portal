-- Runs once when the local Postgres volume is first created.
-- LOCAL DEVELOPMENT ONLY: the password is public in this repo.
CREATE ROLE app LOGIN PASSWORD 'app';
