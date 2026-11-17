-- name: ListActiveSkills :many
-- The picker catalog.
SELECT id, name, is_active
FROM skills
WHERE is_active
ORDER BY sort_order, name;

-- name: GetSkillsByIDs :many
-- Includes inactive skills so the use case can tell "unknown" from "inactive".
SELECT id, name, is_active
FROM skills
WHERE id = ANY(sqlc.arg(ids)::bigint[])
ORDER BY sort_order, name;

-- name: ListEmployeeSkills :many
SELECT s.id, s.name, s.is_active
FROM employee_skills es
JOIN skills s ON s.id = es.skill_id
WHERE es.employee_id = $1
ORDER BY s.sort_order, s.name;

-- name: DeleteEmployeeSkills :exec
DELETE FROM employee_skills WHERE employee_id = $1;

-- name: InsertEmployeeSkills :exec
INSERT INTO employee_skills (employee_id, skill_id)
SELECT sqlc.arg(employee_id), unnest(sqlc.arg(skill_ids)::bigint[]);
