-- name: CreateVerifyEmail :one
INSERT INTO verify_emails (
  username, email, secret_code
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: UpdateVerifyEmail :one
UPDATE verify_emails
SET is_used = $2
WHERE id = $1
RETURNING *;