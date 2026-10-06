-- +goose Up
ALTER TABLE "Country" ADD CONSTRAINT "fk_country_capital" FOREIGN KEY ("capital_id") REFERENCES "City" ("id") ON DELETE SET NULL;
ALTER TABLE "Country" ADD CONSTRAINT "fk_country_image" FOREIGN KEY ("image_id") REFERENCES "StaticPage" ("id") ON DELETE SET NULL;

ALTER TABLE "Holiday" ADD CONSTRAINT "fk_holiday_country" FOREIGN KEY ("country_id") REFERENCES "Country" ("id") ON DELETE CASCADE;
ALTER TABLE "Holiday" ADD CONSTRAINT "fk_holiday_image" FOREIGN KEY ("image_id") REFERENCES "StaticPage" ("id") ON DELETE SET NULL;

ALTER TABLE "City" ADD CONSTRAINT "fk_city_country" FOREIGN KEY ("country_id") REFERENCES "Country" ("id") ON DELETE CASCADE;
ALTER TABLE "City" ADD CONSTRAINT "fk_city_image" FOREIGN KEY ("image_id") REFERENCES "StaticPage" ("id") ON DELETE SET NULL;

ALTER TABLE "TransportNode" ADD CONSTRAINT "fk_transportnode_city" FOREIGN KEY ("city_id") REFERENCES "City" ("id") ON DELETE CASCADE;
ALTER TABLE "TransportNode" ADD CONSTRAINT "fk_transportnode_image" FOREIGN KEY ("image_id") REFERENCES "StaticPage" ("id") ON DELETE SET NULL;

ALTER TABLE "CityDistrict" ADD CONSTRAINT "fk_citydistrict_city" FOREIGN KEY ("city_id") REFERENCES "City" ("id") ON DELETE CASCADE;
ALTER TABLE "CityDistrict" ADD CONSTRAINT "fk_citydistrict_image" FOREIGN KEY ("image_id") REFERENCES "StaticPage" ("id") ON DELETE SET NULL;

ALTER TABLE "Place" ADD CONSTRAINT "fk_place_city" FOREIGN KEY ("city_id") REFERENCES "City" ("id") ON DELETE CASCADE;
ALTER TABLE "Place" ADD CONSTRAINT "fk_place_district" FOREIGN KEY ("district_id") REFERENCES "CityDistrict" ("id") ON DELETE SET NULL;
ALTER TABLE "Place" ADD CONSTRAINT "fk_place_image" FOREIGN KEY ("image_id") REFERENCES "StaticPage" ("id") ON DELETE SET NULL;

ALTER TABLE "Review" ADD CONSTRAINT "fk_review_user" FOREIGN KEY ("user_id") REFERENCES "User" ("id") ON DELETE CASCADE;
ALTER TABLE "Review" ADD CONSTRAINT "fk_review_place" FOREIGN KEY ("place_id") REFERENCES "Place" ("id") ON DELETE CASCADE;
ALTER TABLE "Review" ADD CONSTRAINT "fk_review_image" FOREIGN KEY ("image_id") REFERENCES "StaticPage" ("id") ON DELETE SET NULL;

ALTER TABLE "Trip" ADD CONSTRAINT "fk_trip_user" FOREIGN KEY ("user_id") REFERENCES "User" ("id") ON DELETE CASCADE;
ALTER TABLE "Trip" ADD CONSTRAINT "fk_trip_image" FOREIGN KEY ("image_id") REFERENCES "StaticPage" ("id") ON DELETE SET NULL;

ALTER TABLE "TripPlace" ADD CONSTRAINT "fk_tripplace_trip" FOREIGN KEY ("trip_id") REFERENCES "Trip" ("id") ON DELETE CASCADE;
ALTER TABLE "TripPlace" ADD CONSTRAINT "fk_tripplace_place" FOREIGN KEY ("place_id") REFERENCES "Place" ("id") ON DELETE CASCADE;

-- +goose Down
ALTER TABLE "TripPlace" DROP CONSTRAINT IF EXISTS "fk_tripplace_trip";
ALTER TABLE "TripPlace" DROP CONSTRAINT IF EXISTS "fk_tripplace_place";
ALTER TABLE "Trip" DROP CONSTRAINT IF EXISTS "fk_trip_user";
ALTER TABLE "Trip" DROP CONSTRAINT IF EXISTS "fk_trip_image";
ALTER TABLE "Review" DROP CONSTRAINT IF EXISTS "fk_review_user";
ALTER TABLE "Review" DROP CONSTRAINT IF EXISTS "fk_review_place";
ALTER TABLE "Review" DROP CONSTRAINT IF EXISTS "fk_review_image";
ALTER TABLE "Place" DROP CONSTRAINT IF EXISTS "fk_place_city";
ALTER TABLE "Place" DROP CONSTRAINT IF EXISTS "fk_place_district";
ALTER TABLE "Place" DROP CONSTRAINT IF EXISTS "fk_place_image";
ALTER TABLE "CityDistrict" DROP CONSTRAINT IF EXISTS "fk_citydistrict_city";
ALTER TABLE "CityDistrict" DROP CONSTRAINT IF EXISTS "fk_citydistrict_image";
ALTER TABLE "TransportNode" DROP CONSTRAINT IF EXISTS "fk_transportnode_city";
ALTER TABLE "TransportNode" DROP CONSTRAINT IF EXISTS "fk_transportnode_image";
ALTER TABLE "City" DROP CONSTRAINT IF EXISTS "fk_city_country";
ALTER TABLE "City" DROP CONSTRAINT IF EXISTS "fk_city_image";
ALTER TABLE "Holiday" DROP CONSTRAINT IF EXISTS "fk_holiday_country";
ALTER TABLE "Holiday" DROP CONSTRAINT IF EXISTS "fk_holiday_image";
ALTER TABLE "Country" DROP CONSTRAINT IF EXISTS "fk_country_capital";
ALTER TABLE "Country" DROP CONSTRAINT IF EXISTS "fk_country_image";