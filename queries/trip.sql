-- name: SaveTrip :one
INSERT INTO "Trip" (
    "title",
    "start_date",
    "end_date",
    "budget",
    "status",
    "notes",
    "user_id",
    "image_id"
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING "id";

-- name: FindTripByID :one
SELECT "id", "title", "start_date", "end_date", "budget",
       "status", "notes", "created_at", "user_id", "image_id"
FROM "Trip"
WHERE "id" = $1
LIMIT 1;

-- name: FindTripsByUser :many
SELECT "id", "title", "start_date", "end_date", "budget",
       "status", "notes", "created_at", "user_id", "image_id"
FROM "Trip"
WHERE "user_id" = $1
ORDER BY "start_date" DESC;

-- name: UpdateTrip :exec
UPDATE "Trip"
SET
    "title" = $2,
    "start_date" = $3,
    "end_date" = $4,
    "budget" = $5,
    "status" = $6,
    "notes" = $7,
    "image_id" = $8
WHERE "id" = $1;

-- name: DeleteTrip :exec
DELETE FROM "Trip"
WHERE "id" = $1;

-- =============================================
-- TripPlace (связь поездка – достопримечательность)
-- =============================================

-- name: AddTripPlace :one
INSERT INTO "TripPlace" (
    "day_number",
    "arrival_time",
    "duration_min",
    "notes",
    "visit_status",
    "actual_cost",
    "trip_id",
    "place_id"
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING "id";

-- name: RemoveTripPlace :exec
DELETE FROM "TripPlace"
WHERE "trip_id" = $1 AND "place_id" = $2;

-- name: GetTripPlaces :many
SELECT "id", "day_number", "arrival_time", "duration_min",
       "notes", "visit_status", "actual_cost", "trip_id", "place_id"
FROM "TripPlace"
WHERE "trip_id" = $1
ORDER BY "day_number", "arrival_time";

-- name: UpdateTripPlace :exec
UPDATE "TripPlace"
SET
    "day_number" = $2,
    "arrival_time" = $3,
    "duration_min" = $4,
    "notes" = $5,
    "visit_status" = $6,
    "actual_cost" = $7
WHERE "id" = $1;

-- name: GetTripStatisticsReport :many
SELECT 
    t.id,
    t.title,
    t.user_id,
    t.start_date,
    t.end_date,
    COALESCE(tp.places_count, 0)::bigint AS places_count,
    COALESCE(tp.total_cost, 0)::numeric AS total_cost
FROM "Trip" t
LEFT JOIN (
    SELECT trip_id, COUNT(*) AS places_count, SUM(COALESCE(actual_cost, 0)) AS total_cost
    FROM "TripPlace"
    GROUP BY trip_id
) tp ON tp.trip_id = t.id
WHERE (sqlc.narg('date_from')::text IS NULL OR sqlc.narg('date_from')::text = '' OR t.start_date >= sqlc.narg('date_from')::timestamp)
  AND (sqlc.narg('date_to')::text IS NULL OR sqlc.narg('date_to')::text = '' OR t.end_date <= sqlc.narg('date_to')::timestamp)
ORDER BY t.start_date DESC
LIMIT sqlc.arg('limit')::int;
