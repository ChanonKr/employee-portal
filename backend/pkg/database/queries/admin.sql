-- Admin → Employees and Dashboard. Search patterns are escaped in Go (\, %, _).

-- name: AdminListEmployees :many
SELECT e.id, u.employee_code, e.first_name, e.last_name, e.photo_path, u.role, u.is_active
FROM employees e
JOIN users u ON u.id = e.user_id
WHERE sqlc.narg(pattern)::text IS NULL
   OR u.employee_code ILIKE sqlc.narg(pattern)::text
   OR (e.first_name || ' ' || e.last_name) ILIKE sqlc.narg(pattern)::text
ORDER BY e.first_name, e.last_name, e.id
LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;

-- name: AdminCountEmployees :one
SELECT count(*)
FROM employees e
JOIN users u ON u.id = e.user_id
WHERE sqlc.narg(pattern)::text IS NULL
   OR u.employee_code ILIKE sqlc.narg(pattern)::text
   OR (e.first_name || ' ' || e.last_name) ILIKE sqlc.narg(pattern)::text;

-- name: AdminGetEmployee :one
SELECT sqlc.embed(e), u.employee_code, u.role, u.is_active
FROM employees e
JOIN users u ON u.id = e.user_id
WHERE e.id = $1;

-- name: AdminCountProfiles :one
-- "Complete" matches the profile checklist: photo, all contact fields, at least one skill.
SELECT
  count(*) AS total,
  count(*) FILTER (
    WHERE e.photo_path IS NOT NULL
      AND e.mobile_phone IS NOT NULL
      AND e.contact_email IS NOT NULL
      AND e.address IS NOT NULL
      AND EXISTS (SELECT 1 FROM employee_skills es WHERE es.employee_id = e.id)
  ) AS complete
FROM employees e;

-- name: AdminTopSkills :many
SELECT s.id, s.name, count(*) AS used_by
FROM employee_skills es
JOIN skills s ON s.id = es.skill_id
GROUP BY s.id
ORDER BY used_by DESC, s.name
LIMIT $1;

-- name: AdminRecentChanges :many
SELECT l.id, l.employee_id, u.employee_code, e.first_name, e.last_name, l.field, l.changed_at,
       cb.role AS changed_by_role
FROM profile_change_logs l
JOIN employees e ON e.id = l.employee_id
JOIN users u ON u.id = e.user_id
JOIN users cb ON cb.id = l.changed_by
ORDER BY l.changed_at DESC, l.id DESC
LIMIT $1;
