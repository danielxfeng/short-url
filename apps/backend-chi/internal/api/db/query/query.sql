-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: UpsertUser :one
INSERT INTO users (
  provider, provider_id, display_name, profile_pic
) VALUES (
  $1, $2, $3, $4
) ON CONFLICT (provider, provider_id) DO UPDATE
  SET display_name = EXCLUDED.display_name,
      profile_pic = EXCLUDED.profile_pic
RETURNING *;

-- name: DeleteUser :one
DELETE FROM users
WHERE id = $1
RETURNING *;

-- name: CreateLink :one
INSERT INTO links (
  user_id, code, original_url
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetLinkByCode :one
SELECT * FROM links
WHERE code = $1 AND deleted_at IS NULL;

-- name: GetLinksByUserID :many
SELECT * FROM links
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: SetLinkDeleted :one
UPDATE links
SET deleted_at = now()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id;

-- name: SetLinkClicked :one
UPDATE links
SET clicks = clicks + 1
WHERE id = $1 AND deleted_at IS NULL
RETURNING clicks;

-- name: ResetDb :exec
TRUNCATE users, links RESTART IDENTITY CASCADE;
