-- name: GetSetting :one
SELECT value FROM app_settings WHERE key = $1;

-- name: GetSettingDetail :one
SELECT s.value, s.updated_at, u.id AS updated_by_id, u.employee_code AS updated_by_code
FROM app_settings s
LEFT JOIN users u ON u.id = s.updated_by
WHERE s.key = $1;

-- name: UpsertSetting :exec
INSERT INTO app_settings (key, value, updated_by, updated_at)
VALUES ($1, $2, $3, now())
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value, updated_by = EXCLUDED.updated_by, updated_at = now();
