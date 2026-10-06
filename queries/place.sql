-- name: SavePlace :one
INSERT INTO "Place" (
    "name",
    "category",
    "description",
    "coordinates",
    "address", 
    "opening_hours",
    "price_info",
    "avg_visit_duration_min",
    "avg_rating",
    "reviews_count",
    "contact_phone",
    "website",
    "city_id",
    "district_id",
    "image_id"
) VALUES
($1, $2, 
$3, $4, 
$5, $6, 
$7, $8, 
$9, $10, 
$11, $12, 
$13, $14, $15)
RETURNING "id";


-- name: FindPlaceByID :one
SELECT "id", "name",
    "category", "description",
    "coordinates", "address", 
    "opening_hours", "price_info",
    "avg_visit_duration_min", "avg_rating", "reviews_count",
    "contact_phone", "website",
    "city_id", "district_id",
    "image_id"
FROM "Place"
WHERE "id" = $1
LIMIT 1;


-- name: FindPlacesByCity :many
SELECT "id", "name",
    "category", "description",
    "coordinates", "address", 
    "opening_hours", "price_info",
    "avg_visit_duration_min", "avg_rating", "reviews_count",
    "contact_phone", "website",
    "city_id", "district_id",
    "image_id"
FROM "Place"
WHERE "city_id" = $1;

-- name: FindPlacesByCategory :many
SELECT "id", "name",
    "category", "description",
    "coordinates", "address", 
    "opening_hours", "price_info",
    "avg_visit_duration_min", "avg_rating", "reviews_count",
    "contact_phone", "website",
    "city_id", "district_id",
    "image_id"
FROM "Place"
WHERE "category" = $1;

-- name: UpdatePlace :exec
UPDATE "Place" 
SET "name" = $2,
    "category" = $3,
    "description" = $4,
    "coordinates" = $5,
    "address" = $6,
    "opening_hours" = $7,
    "price_info" = $8,
    "avg_visit_duration_min" = $9,
    "avg_rating" = $10,
    "reviews_count" = $11,
    "contact_phone" = $12,
    "website" = $13,
    "city_id" = $14,
    "district_id" = $15,
    "image_id" = $16
WHERE "id" = $1;

-- name: DeletePlace :exec
DELETE FROM "Place"
WHERE "id" = $1;

-- name: UpdateRating :exec
UPDATE "Place" 
SET 
    "avg_rating" = $2,
    "reviews_count" = $3
WHERE "id" = $1;

-- name: SaveReview :one
INSERT INTO "Review" (
    "rating",
    "comment",
    "visit_date",
    "user_id",
    "place_id",
    "image_id"
) VALUES
($1, $2, 
$3, $4, 
$5, $6)
RETURNING "id";


-- name: FindReviewByID :one
SELECT "id", "rating", "comment", "visit_date", "created_at",
       "is_moderated", "is_approved", "moderation_comment",
       "user_id", "place_id", "image_id"
FROM "Review"
WHERE "id" = $1
LIMIT 1;

-- name: FindReviewsByUserAndPlace :many
SELECT "id", "rating", "comment", "visit_date", "created_at",
       "is_moderated", "is_approved", "moderation_comment",
       "user_id", "place_id", "image_id"
FROM "Review"
WHERE "user_id" = $1 AND "place_id" = $2 AND is_approved = true;

-- name: FindReviewsByPlace :many
SELECT "id", "rating", "comment", "visit_date", "created_at",
       "is_moderated", "is_approved", "moderation_comment",
       "user_id", "place_id", "image_id"
FROM "Review"
WHERE "place_id" = $1 AND is_approved = true;

-- name: FindReviewsByUser :many
SELECT "id", "rating", "comment", "visit_date", "created_at",
       "is_moderated", "is_approved", "moderation_comment",
       "user_id", "place_id", "image_id"
FROM "Review"
WHERE "user_id" = $1;

-- name: UpdateReview :exec
UPDATE "Review" 
SET "rating" = $2,
    "comment" = $3,
    "visit_date" = $4,
    "image_id" = $5
WHERE "id" = $1;


-- name: DeleteReview :exec
DELETE FROM "Review"
WHERE "id" = $1;

-- name: ModerateReview :exec
UPDATE "Review" 
SET 
    "is_moderated" = true,
    "is_approved" = $2,
    "moderation_comment" = $3
WHERE "id" = $1;

-- name: GetPopularPlacesReport :many
SELECT 
    p.id,
    p.name,
    p.category,
    COALESCE(tp.trips_count, 0)::bigint AS trips_count,
    COALESCE(p.reviews_count, 0)::bigint AS reviews_count,
    COALESCE(p.avg_rating, 0)::numeric AS avg_rating
FROM "Place" p
LEFT JOIN (
    SELECT place_id, COUNT(*) AS trips_count
    FROM "TripPlace"
    GROUP BY place_id
) tp ON tp.place_id = p.id
ORDER BY trips_count DESC, p.reviews_count DESC
LIMIT sqlc.arg('limit')::int;

-- name: FindPendingReviews :many
SELECT 
    r.id, r.rating, r.comment, r.visit_date, r.created_at,
    r.is_moderated, r.is_approved, r.moderation_comment,
    r.user_id, r.place_id, r.image_id,
    u.username
FROM "Review" r
JOIN "User" u ON u.id = r.user_id
WHERE r.is_moderated = false
ORDER BY r.created_at ASC
LIMIT $1 OFFSET $2;
