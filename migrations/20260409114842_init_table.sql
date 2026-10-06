-- +goose Up
CREATE TABLE "StaticPage" (
  "id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
  "slug" VARCHAR(100) UNIQUE NOT NULL,
  "title" VARCHAR(255) NOT NULL,
  "content" TEXT,
  "meta_description" VARCHAR(255),
  "image_url" VARCHAR(500),
  "image_alt" VARCHAR(255),
  "updated_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP),
  "published_at" TIMESTAMP,
  "is_published" BOOLEAN DEFAULT false
);

CREATE TABLE "Country" (
  "id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
  "name" VARCHAR(255) NOT NULL,
  "area" NUMERIC(12,2),
  "population" BIGINT,
  "gdp" NUMERIC(20,2),
  "currency" VARCHAR(3),
  "visa_requirements" TEXT,
  "description" TEXT,
  "safety_tips" TEXT,
  "best_season" VARCHAR(50),
  "language" VARCHAR(100),
  "phone_code" VARCHAR(10),
  "religion" VARCHAR(100),
  "capital_id" UUID,
  "image_id" UUID
);

CREATE TABLE "Holiday" (
  "id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
  "name" VARCHAR(255) NOT NULL,
  "date" DATE NOT NULL,
  "description" TEXT,
  "traditions" TEXT,
  "history" TEXT,
  "is_national" BOOLEAN DEFAULT false,
  "country_id" UUID NOT NULL,
  "image_id" UUID
);

CREATE TABLE "City" (
  "id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
  "name" VARCHAR(255) NOT NULL,
  "population" BIGINT,
  "is_capital" BOOLEAN DEFAULT false,
  "coordinates" POINT,
  "description" TEXT,
  "timezone" VARCHAR(50),
  "travel_tips" TEXT,
  "country_id" UUID NOT NULL,
  "image_id" UUID
);

CREATE TABLE "TransportNode" (
  "id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
  "name" VARCHAR(255) NOT NULL,
  "type" VARCHAR(50) NOT NULL,
  "coordinates" POINT,
  "address" TEXT,
  "city_id" UUID NOT NULL,
  "image_id" UUID
);

CREATE TABLE "CityDistrict" (
  "id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
  "name" VARCHAR(255) NOT NULL,
  "description" TEXT,
  "coordinates" POINT,
  "city_id" UUID NOT NULL,
  "image_id" UUID
);

CREATE TABLE "Place" (
  "id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
  "name" VARCHAR(255) NOT NULL,
  "category" VARCHAR(50) NOT NULL,
  "description" TEXT,
  "coordinates" POINT,
  "address" TEXT,
  "opening_hours" TEXT,
  "price_info" TEXT,
  "avg_visit_duration_min" INTEGER,
  "avg_rating" NUMERIC(3,2) DEFAULT 0,
  "reviews_count" INTEGER DEFAULT 0,
  "contact_phone" VARCHAR(50),
  "website" VARCHAR(255),
  "city_id" UUID NOT NULL,
  "district_id" UUID,
  "image_id" UUID
);

CREATE TABLE "User" (
  "id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
  "role" VARCHAR(20) NOT NULL DEFAULT 'USER' CHECK (role IN ('USER', 'MODERATOR', 'ANALYST')),
  "username" VARCHAR(100) UNIQUE NOT NULL,
  "email" VARCHAR(255) UNIQUE NOT NULL,
  "password_hash" VARCHAR(255) NOT NULL,
  "phone" VARCHAR(50),
  "registered_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP),
  "birth_date" DATE,
  "country_of_residence" VARCHAR(100),
  "avatar_url" VARCHAR(500),
  "favorite_categories" TEXT[],
  "is_blocked" BOOLEAN DEFAULT false,
  "blocked_at" TIMESTAMP,
  "block_reason" TEXT
);

CREATE TABLE "Review" (
  "id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
  "rating" INTEGER NOT NULL,
  "comment" TEXT,
  "visit_date" DATE NOT NULL,
  "created_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP),
  "is_moderated" BOOLEAN DEFAULT false,
  "is_approved" BOOLEAN DEFAULT false,
  "moderation_comment" TEXT,
  "user_id" UUID NOT NULL,
  "place_id" UUID NOT NULL,
  "image_id" UUID
);

CREATE TABLE "Trip" (
  "id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
  "title" VARCHAR(255) NOT NULL,
  "start_date" DATE NOT NULL,
  "end_date" DATE NOT NULL,
  "budget" NUMERIC(10,2),
  "status" VARCHAR(50) DEFAULT 'DRAFT',
  "notes" TEXT,
  "created_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP),
  "user_id" UUID NOT NULL,
  "image_id" UUID
);

CREATE TABLE "TripPlace" (
  "id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
  "day_number" INTEGER NOT NULL,
  "arrival_time" TIME,
  "duration_min" INTEGER,
  "notes" TEXT,
  "visit_status" VARCHAR(50) DEFAULT 'PLANNED',
  "actual_cost" NUMERIC(10,2),
  "trip_id" UUID NOT NULL,
  "place_id" UUID NOT NULL
);

-- +goose Down

DROP TABLE IF EXISTS "TripPlace" CASCADE;
DROP TABLE IF EXISTS "Trip" CASCADE;
DROP TABLE IF EXISTS "Review" CASCADE;
DROP TABLE IF EXISTS "User" CASCADE;
DROP TABLE IF EXISTS "Place" CASCADE;
DROP TABLE IF EXISTS "CityDistrict" CASCADE;
DROP TABLE IF EXISTS "TransportNode" CASCADE;
DROP TABLE IF EXISTS "City" CASCADE;
DROP TABLE IF EXISTS "Holiday" CASCADE;
DROP TABLE IF EXISTS "Country" CASCADE;
DROP TABLE IF EXISTS "StaticPage" CASCADE;
