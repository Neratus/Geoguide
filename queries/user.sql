-- name: SaveUser :one
INSERT INTO "User" (
    "role",
    "username",
    "email",
    "password_hash",
    "phone",
    "birth_date",
    "country_of_residence",
    "avatar_url",
    "favorite_categories",
    "email_verified",
    "phone_verified",
    "two_factor_enabled",
    "two_factor_secret",
    "backup_codes"
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, false, false, false, NULL, NULL)
RETURNING "id";

-- name: FindUserByID :one
SELECT "id", "role", "username", "email", "password_hash",
       "phone", "registered_at", "birth_date", "country_of_residence",
       "avatar_url", "favorite_categories",
       "is_blocked", "blocked_at", "block_reason",
       "email_verified", "phone_verified", "two_factor_enabled",
       "two_factor_secret", "backup_codes"
FROM "User"
WHERE "id" = $1
LIMIT 1;

-- name: FindUserByUsername :one
SELECT "id", "role", "username", "email", "password_hash",
       "phone", "registered_at", "birth_date", "country_of_residence",
       "avatar_url", "favorite_categories",
       "is_blocked", "blocked_at", "block_reason",
       "email_verified", "phone_verified", "two_factor_enabled",
       "two_factor_secret", "backup_codes"
FROM "User"
WHERE "username" = $1
LIMIT 1;

-- name: FindUserByEmail :one
SELECT "id", "role", "username", "email", "password_hash",
       "phone", "registered_at", "birth_date", "country_of_residence",
       "avatar_url", "favorite_categories",
       "is_blocked", "blocked_at", "block_reason",
       "email_verified", "phone_verified", "two_factor_enabled",
       "two_factor_secret", "backup_codes"
FROM "User"
WHERE "email" = $1
LIMIT 1;

-- name: UpdateUser :exec
UPDATE "User"
SET
    "role" = $2,
    "username" = $3,
    "email" = $4,
    "password_hash" = $5,
    "phone" = $6,
    "birth_date" = $7,
    "country_of_residence" = $8,
    "avatar_url" = $9,
    "favorite_categories" = $10
WHERE "id" = $1;

-- name: DeleteUser :exec
DELETE FROM "User"
WHERE "id" = $1;

-- name: BlockUser :exec
UPDATE "User"
SET
    "is_blocked" = true,
    "blocked_at" = CURRENT_TIMESTAMP,
    "block_reason" = $2
WHERE "id" = $1;

-- name: UnblockUser :exec
UPDATE "User"
SET
    "is_blocked" = false,
    "blocked_at" = NULL,
    "block_reason" = NULL
WHERE "id" = $1;

-- name: UpdateEmailVerified :exec
UPDATE "User"
SET "email_verified" = $2
WHERE "id" = $1;

-- name: UpdatePhoneVerified :exec
UPDATE "User"
SET "phone_verified" = $2
WHERE "id" = $1;

-- name: EnableTwoFactor :exec
UPDATE "User"
SET "two_factor_enabled" = true,
    "two_factor_secret" = $2,
    "backup_codes" = $3
WHERE "id" = $1;

-- name: DisableTwoFactor :exec
UPDATE "User"
SET "two_factor_enabled" = false,
    "two_factor_secret" = NULL,
    "backup_codes" = NULL
WHERE "id" = $1;

-- name: GetUserActivityReport :many
SELECT 
    u.id,
    u.username,
    COALESCE(t.trips_created, 0)::bigint AS trips_created,
    COALESCE(r.reviews_written, 0)::bigint AS reviews_written,
    GREATEST(
        COALESCE(t.last_trip, '1970-01-01'::timestamp), 
        COALESCE(r.last_review, '1970-01-01'::timestamp)
    ) AS last_active
FROM "User" u
LEFT JOIN (
    SELECT user_id, COUNT(*) AS trips_created, MAX(created_at) AS last_trip
    FROM "Trip"
    WHERE (sqlc.narg('date_from')::text IS NULL OR sqlc.narg('date_from')::text = '' OR created_at >= sqlc.narg('date_from')::timestamp)
      AND (sqlc.narg('date_to')::text IS NULL OR sqlc.narg('date_to')::text = '' OR created_at <= sqlc.narg('date_to')::timestamp)
    GROUP BY user_id
) t ON t.user_id = u.id
LEFT JOIN (
    SELECT user_id, COUNT(*) AS reviews_written, MAX(created_at) AS last_review
    FROM "Review"
    WHERE (sqlc.narg('date_from')::text IS NULL OR sqlc.narg('date_from')::text = '' OR created_at >= sqlc.narg('date_from')::timestamp)
      AND (sqlc.narg('date_to')::text IS NULL OR sqlc.narg('date_to')::text = '' OR created_at <= sqlc.narg('date_to')::timestamp)
    GROUP BY user_id
) r ON r.user_id = u.id
ORDER BY last_active DESC
LIMIT sqlc.arg('limit')::int;

-- name: FindAllUsers :many
SELECT "id", "role", "username", "email", "password_hash", "phone", "registered_at", "birth_date",
       "country_of_residence", "avatar_url", "favorite_categories",
       "is_blocked", "blocked_at", "block_reason",
       "email_verified", "phone_verified", "two_factor_enabled", "two_factor_secret", "backup_codes"
FROM "User"
WHERE ($1::text = '' OR username ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%')
ORDER BY registered_at DESC
LIMIT $2 OFFSET $3;
