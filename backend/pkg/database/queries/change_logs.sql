-- name: InsertChangeLog :exec
INSERT INTO profile_change_logs (employee_id, field, old_value, new_value, changed_by)
VALUES ($1, $2, $3, $4, $5);

-- name: ListHistory :many
-- days NULL = no time window (admin). field NULL = all fields.
-- Fetch page_limit = pageSize + 1 rows to know whether a next page exists.
SELECT l.id, l.field, l.old_value, l.new_value, l.changed_at, l.changed_by, u.role AS changed_by_role
FROM profile_change_logs l
JOIN users u ON u.id = l.changed_by
WHERE l.employee_id = sqlc.arg(employee_id)
  AND (sqlc.narg(days)::int IS NULL OR l.changed_at >= now() - make_interval(days => sqlc.narg(days)::int))
  AND (sqlc.narg(field)::text IS NULL OR l.field = sqlc.narg(field)::text)
ORDER BY l.changed_at DESC, l.id DESC
LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;

-- name: HasOlderHistory :one
-- Drives the empty state: "No changes yet" vs "No changes in the last N days".
SELECT EXISTS (
  SELECT 1 FROM profile_change_logs
  WHERE employee_id = sqlc.arg(employee_id)
    AND changed_at < now() - make_interval(days => sqlc.arg(days)::int)
);
