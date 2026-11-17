-- name: ListEmployeeSkills :many
SELECT s.id, s.name
FROM employee_skills es
JOIN skills s ON s.id = es.skill_id
WHERE es.employee_id = $1
ORDER BY s.sort_order, s.name;
