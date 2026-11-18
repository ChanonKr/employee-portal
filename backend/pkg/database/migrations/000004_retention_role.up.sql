-- The retention job connects as role "retention": it may delete old change history,
-- which the API role "app" can never do (the log is append-only for the app).
-- LOGIN and the password are set outside migrations (db-init for local dev, secret
-- manager in prod), as for "app".
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'retention') THEN
    CREATE ROLE retention NOLOGIN;
  END IF;
END
$$;

GRANT USAGE ON SCHEMA public TO retention;
GRANT SELECT, DELETE ON profile_change_logs TO retention;
GRANT SELECT ON employees TO retention; -- current photos must not be deleted
