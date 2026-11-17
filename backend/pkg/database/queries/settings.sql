-- name: GetSetting :one
SELECT value FROM app_settings WHERE key = $1;
