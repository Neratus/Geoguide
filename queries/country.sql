-- name: SaveCountry :one
INSERT INTO "Country" (
    "name",
    "area",
    "population",
    "gdp",
    "currency",
    "visa_requirements",
    "description",
    "safety_tips",
    "best_season",
    "language",
    "phone_code",
    "religion",
    "capital_id",
    "image_id"
) VALUES
($1, $2, 
$3, $4, 
$5, $6, 
$7, $8, 
$9, $10, 
$11, $12, 
$13, $14)
RETURNING "id";

-- name: FindCountryByID :one
SELECT "id", "name", "area", "population",
       "gdp", "currency", "visa_requirements",
       "description", "safety_tips", "best_season",
       "language", "phone_code", "religion",
       "capital_id", "image_id"
FROM "Country"
WHERE "id" = $1
LIMIT 1;

-- name: GetCountries :many
SELECT "id", "name", "area", "population",
       "gdp", "currency", "visa_requirements",
       "description", "safety_tips", "best_season",
       "language", "phone_code", "religion",
       "capital_id", "image_id"
FROM "Country"
LIMIT 200;

-- name: FindCountryByName :one
SELECT "id", "name", "area", "population",
       "gdp", "currency", "visa_requirements",
       "description", "safety_tips", "best_season",
       "language", "phone_code", "religion",
       "capital_id", "image_id"
FROM "Country"
WHERE "name" = $1
LIMIT 1;

-- name: UpdateCountry :exec
UPDATE "Country"
SET 
    "name" = $2,
    "area" = $3,
    "population" = $4,
    "gdp"  = $5,
    "currency" = $6,
    "visa_requirements" = $7,
    "description" = $8,
    "safety_tips" = $9,
    "best_season" = $10,
    "language" = $11,
    "phone_code" = $12,
    "religion" = $13,
    "capital_id" = $14,
    "image_id" = $15
WHERE "id" = $1;

-- name: DeleteCountry :exec
DELETE FROM "Country"
WHERE "id" = $1;

-- name: SaveHoliday :one
INSERT INTO "Holiday" (
    "name",
    "date",
    "description",
    "traditions",
    "history",
    "is_national",
    "country_id",
    "image_id"
) VALUES
($1, $2, 
$3, $4, 
$5, $6, 
$7, $8)
RETURNING "id";

-- name: FindHolidayByID :one
SELECT "id", "name", "date",
    "description", "traditions",
    "history", "is_national",
    "country_id", "image_id"
FROM "Holiday"
WHERE "id" = $1
LIMIT 1;


-- name: FindHolidaysByCountry :many
SELECT "id", "name", "date",
    "description", "traditions",
    "history", "is_national",
    "country_id", "image_id"
FROM "Holiday"
WHERE "country_id" = $1
LIMIT 30;


-- name: FindHolidaysByDate :many
SELECT "id", "name", "date",
    "description", "traditions",
    "history", "is_national",
    "country_id", "image_id"
FROM "Holiday"
WHERE "date" = $1;

-- name: UpdateHoliday :exec
UPDATE "Holiday"
SET 
    "name" = $2,
    "date" = $3,
    "description" = $4,
    "traditions" = $5,
    "history" = $6,
    "is_national" = $7,
    "country_id" = $8,
    "image_id" = $9
WHERE "id" = $1;


-- name: DeleteHoliday :exec
DELETE FROM "Holiday"
WHERE "id" = $1;
