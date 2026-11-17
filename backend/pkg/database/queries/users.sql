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

-- name: CreateUser :one
-- New accounts must change the temporary password on first sign-in.
INSERT INTO users (employee_code, password_hash, role, must_change_password)
VALUES ($1, $2, $3, TRUE)
RETURNING id;

-- name: CreateEmployee :one
INSERT INTO employees (user_id, first_name, last_name)
VALUES ($1, $2, $3)
RETURNING id;

-- name: SetPassword :execrows
UPDATE users SET password_hash = $2, must_change_password = $3
WHERE id = $1;

-- name: GetUserIDByEmployee :one
SELECT user_id FROM employees WHERE id = $1;
