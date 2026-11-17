-- name: GetProfile :one
SELECT sqlc.embed(e), u.employee_code
FROM employees e
JOIN users u ON u.id = e.user_id
WHERE e.id = $1;

-- name: GetProfileForUpdate :one
-- Locks the row until the transaction ends so concurrent edits are serialized.
SELECT * FROM employees
WHERE id = $1
FOR UPDATE;

-- name: UpdateContactField :one
-- The column name comes from a whitelist in the repository, never from user input.
UPDATE employees SET
  mobile_phone  = CASE WHEN sqlc.arg(field)::text = 'mobile_phone'  THEN sqlc.narg(value)::text ELSE mobile_phone  END,
  contact_email = CASE WHEN sqlc.arg(field)::text = 'contact_email' THEN sqlc.narg(value)::text ELSE contact_email END,
  address       = CASE WHEN sqlc.arg(field)::text = 'address'       THEN sqlc.narg(value)::text ELSE address       END,
  version       = version + 1,
  updated_at    = now()
WHERE id = sqlc.arg(id) AND version = sqlc.arg(version)
RETURNING version;
