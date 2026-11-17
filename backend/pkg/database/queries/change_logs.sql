-- name: InsertChangeLog :exec
INSERT INTO profile_change_logs (employee_id, field, old_value, new_value, changed_by)
VALUES ($1, $2, $3, $4, $5);
