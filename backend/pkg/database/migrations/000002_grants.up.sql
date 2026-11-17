-- The API connects as role "app". Its password and LOGIN are set outside migrations
-- (db-init for local dev, secret manager in prod). Create it NOLOGIN here so the grants
-- below also work on a fresh database such as the integration-test container.
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app') THEN
    CREATE ROLE app NOLOGIN;
  END IF;
END
$$;

GRANT USAGE ON SCHEMA public TO app;
GRANT SELECT, INSERT, UPDATE, DELETE ON users, employees, skills, employee_skills, app_settings TO app;

-- Change history is append-only: no UPDATE/DELETE even if the app has a bug.
GRANT SELECT, INSERT ON profile_change_logs TO app;

GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO app;
