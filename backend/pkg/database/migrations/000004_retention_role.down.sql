REVOKE ALL ON profile_change_logs, employees FROM retention;
REVOKE USAGE ON SCHEMA public FROM retention;
DROP ROLE IF EXISTS retention;
