-- name: SaveStatic :one
INSERT INTO "StaticPage" (
    "id",
    "slug",
    "title",
    "content",
    "meta_description",
    "image_url",
    "image_alt",
    "is_published"
) VALUES
($1, $2, 
$3, $4, 
$5, $6, 
$7, $8)
RETURNING "id";

-- name: FindStaticByID :one
SELECT "id", "slug", "title", "content", "meta_description",
       "image_url", "image_alt", "updated_at", "published_at", "is_published"
FROM "StaticPage"
WHERE "id" = $1
LIMIT 1;

-- name: FindStaticPageBySlug :one
SELECT "id", "slug", "title", "content", "meta_description",
       "image_url", "image_alt", "updated_at", "published_at", "is_published"
FROM "StaticPage" 
WHERE "slug" = $1 
LIMIT 1;

-- name: UpdateStatic :exec
UPDATE "StaticPage" 
SET "slug" = $2,
    "title" = $3,
    "content" = $4,
    "meta_description" = $5,
    "image_url" = $6,
    "image_alt" = $7,
    "is_published" = $8
WHERE "id" = $1;


-- name: DeleteStatic :exec
DELETE FROM "StaticPage"
WHERE "id" = $1;
