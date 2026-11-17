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

-- name: AdminListSkills :many
SELECT s.id, s.name, s.is_active, s.sort_order, count(es.employee_id) AS used_by
FROM skills s
LEFT JOIN employee_skills es ON es.skill_id = s.id
GROUP BY s.id
ORDER BY s.sort_order, s.name;

-- name: CreateSkill :one
-- New skills go to the end of the list.
INSERT INTO skills (name, sort_order)
VALUES ($1, (SELECT COALESCE(MAX(sort_order), 0) + 1 FROM skills))
RETURNING id;

-- name: UpdateSkill :execrows
-- NULL keeps the current value. 0 rows = unknown id.
UPDATE skills SET
  name      = COALESCE(sqlc.narg(name)::text, name),
  is_active = COALESCE(sqlc.narg(is_active)::bool, is_active)
WHERE id = sqlc.arg(id);

-- name: ReorderSkills :execrows
-- ids[i] gets sort_orders[i]. The repository runs it in a transaction and rolls
-- back unless every id matched.
UPDATE skills
SET sort_order = (sqlc.arg(sort_orders)::int[])[array_position(sqlc.arg(ids)::bigint[], id)]
WHERE id = ANY(sqlc.arg(ids)::bigint[]);
