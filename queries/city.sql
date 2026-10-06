-- name: SaveCity :one
INSERT INTO "City" (
    "name",
    "population",
    "is_capital",
    "coordinates",
    "description",
    "timezone",
    "travel_tips",
    "country_id",
    "image_id"
) VALUES
($1, $2, 
$3, $4, 
$5, $6, 
$7, $8, $9)
RETURNING "id";

-- name: FindCityByID :one
SELECT "id", "name", "population", "is_capital", "coordinates",
       "description", "timezone", "travel_tips", "country_id", "image_id"
FROM "City"
WHERE "id" = $1
LIMIT 1;


-- name: FindCitiesByCountry :many
SELECT  "id", "name", "population", "is_capital", "coordinates",
       "description", "timezone", "travel_tips", "country_id", "image_id"
FROM "City"
WHERE "country_id" = $1;


-- name: FindCityByName :one
SELECT  "id", "name", "population", "is_capital", "coordinates",
       "description", "timezone", "travel_tips", "country_id", "image_id"
FROM "City"
WHERE "name" = $1
LIMIT 1;

-- name: UpdateCity :exec
UPDATE "City"
SET 
    "name" = $2,
    "population" = $3,
    "is_capital" = $4,
    "coordinates" = $5,
    "description" = $6,
    "timezone" = $7,
    "travel_tips" = $8,
    "country_id" = $9,
    "image_id" = $10
WHERE "id" = $1;

-- name: DeleteCity :exec
DELETE FROM "City"
WHERE "id" = $1;

-- name: SaveDistrict :one
INSERT INTO "CityDistrict" (
    "name",
    "coordinates",
    "description",
    "city_id",
    "image_id"
) VALUES
($1, $2, 
$3, $4, $5)
RETURNING "id";

-- name: FindDistrictsByCity :many
SELECT "id", "name", "description", "coordinates",
    "city_id", "image_id"
FROM "CityDistrict"
WHERE "city_id" = $1;


-- name: UpdateDistrict :exec
UPDATE "CityDistrict"
SET 
    "name" = $2,
    "coordinates" = $3,
    "description" = $4,
    "city_id" = $5,
    "image_id" = $6
WHERE "id" = $1;

-- name: DeleteDistrict :exec
DELETE FROM "CityDistrict"
WHERE "id" = $1;


-- name: SaveTransportNode :one
INSERT INTO "TransportNode" (
    "name",
    "type",
    "coordinates",
    "address",
    "city_id",
    "image_id"
) VALUES
($1, $2, $3, $4, $5, $6)
RETURNING "id";

-- name: FindTransportNodeByID :one
SELECT "id", "name", "type",
    "coordinates", "address",
    "city_id", "image_id"
FROM "TransportNode"
WHERE "id" = $1
LIMIT 1;

-- name: FindTransportNodesByCity :many
SELECT "id", "name", "type",
    "coordinates", "address",
    "city_id", "image_id"
FROM "TransportNode"
WHERE "city_id" = $1;


-- name: UpdateTransportNode :exec
UPDATE "TransportNode"
SET 
    "name" = $2,
    "type" = $3,
    "coordinates" = $4,
    "address" = $5,
    "city_id" = $6,
    "image_id" = $7
WHERE "id" = $1;

-- name: DeleteTransportNode :exec
DELETE FROM "TransportNode"
WHERE "id" = $1;

-- name: SearchCities :many
SELECT "id", "name", "population", "is_capital", "coordinates",
       "description", "timezone", "travel_tips", "country_id", "image_id"
FROM "City"
WHERE ($1 = '' OR name ILIKE '%' || $1 || '%')
ORDER BY name
LIMIT $2 OFFSET $3;
