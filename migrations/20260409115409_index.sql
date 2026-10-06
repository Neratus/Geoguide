-- +goose Up
SELECT 'up SQL query';
CREATE UNIQUE INDEX ON "StaticPage" ("slug");

CREATE INDEX ON "Country" ("name");

CREATE INDEX ON "Country" ("capital_id");

CREATE INDEX ON "Holiday" ("country_id");

CREATE INDEX ON "Holiday" ("date");

CREATE INDEX ON "City" ("country_id");

CREATE INDEX ON "City" ("name");

CREATE INDEX ON "TransportNode" ("city_id");

CREATE INDEX ON "TransportNode" ("type");

CREATE INDEX ON "CityDistrict" ("city_id");

CREATE INDEX ON "CityDistrict" ("name");

CREATE INDEX ON "Place" ("city_id");

CREATE INDEX ON "Place" ("district_id");

CREATE INDEX ON "Place" ("category");

CREATE INDEX ON "Place" ("avg_rating");

CREATE INDEX ON "User" ("username");

CREATE INDEX ON "User" ("email");

CREATE INDEX ON "Review" ("user_id");

CREATE INDEX ON "Review" ("place_id");

CREATE INDEX ON "Review" ("created_at");

CREATE INDEX ON "Trip" ("user_id");

CREATE INDEX ON "Trip" ("start_date", "end_date");

CREATE INDEX ON "TripPlace" ("trip_id");

CREATE INDEX ON "TripPlace" ("place_id");

CREATE INDEX ON "TripPlace" ("trip_id", "day_number");

CREATE UNIQUE INDEX "unique_trip_place_day" ON "TripPlace" ("trip_id", "place_id", "day_number");


-- +goose Down
DROP INDEX IF EXISTS "StaticPage_slug_idx";
DROP INDEX IF EXISTS "Country_name_idx";
DROP INDEX IF EXISTS "Country_capital_id_idx";
DROP INDEX IF EXISTS "Holiday_country_id_idx";
DROP INDEX IF EXISTS "Holiday_date_idx";
DROP INDEX IF EXISTS "City_country_id_idx";
DROP INDEX IF EXISTS "City_name_idx";
DROP INDEX IF EXISTS "TransportNode_city_id_idx";
DROP INDEX IF EXISTS "TransportNode_type_idx";
DROP INDEX IF EXISTS "CityDistrict_city_id_idx";
DROP INDEX IF EXISTS "CityDistrict_name_idx";
DROP INDEX IF EXISTS "Place_city_id_idx";
DROP INDEX IF EXISTS "Place_district_id_idx";
DROP INDEX IF EXISTS "Place_category_idx";
DROP INDEX IF EXISTS "Place_avg_rating_idx";
DROP INDEX IF EXISTS "User_username_idx";
DROP INDEX IF EXISTS "User_email_idx";
DROP INDEX IF EXISTS "Review_user_id_idx";
DROP INDEX IF EXISTS "Review_place_id_idx";
DROP INDEX IF EXISTS "Review_created_at_idx";
DROP INDEX IF EXISTS "Trip_user_id_idx";
DROP INDEX IF EXISTS "Trip_start_date_end_date_idx";
DROP INDEX IF EXISTS "TripPlace_trip_id_idx";
DROP INDEX IF EXISTS "TripPlace_place_id_idx";
DROP INDEX IF EXISTS "TripPlace_trip_id_day_number_idx";
DROP INDEX IF EXISTS "unique_trip_place_day";
