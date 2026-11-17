-- name: GetUserByCode :one
SELECT u.id, u.employee_code, u.password_hash, u.role, u.is_active, u.must_change_password,
       e.id AS employee_id
FROM users u
LEFT JOIN employees e ON e.user_id = u.id
WHERE u.employee_code = $1;

-- name: GetUserByID :one
SELECT u.id, u.employee_code, u.password_hash, u.role, u.is_active, u.must_change_password,
       e.id AS employee_id
FROM users u
LEFT JOIN employees e ON e.user_id = u.id
WHERE u.id = $1;
