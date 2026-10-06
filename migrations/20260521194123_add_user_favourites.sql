-- +goose Up
CREATE TABLE "UserFavoritePlace" (
    "user_id" UUID NOT NULL,
    "place_id" UUID NOT NULL,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("user_id", "place_id")
);

CREATE INDEX idx_user_favourite_place_user ON "UserFavoritePlace" ("user_id");
CREATE INDEX idx_user_favourite_place_place ON "UserFavoritePlace" ("place_id");

ALTER TABLE "UserFavoritePlace"
    ADD CONSTRAINT fk_user_favourite_user FOREIGN KEY ("user_id") REFERENCES "User" ("id") ON DELETE CASCADE;
ALTER TABLE "UserFavoritePlace"
    ADD CONSTRAINT fk_user_favourite_place FOREIGN KEY ("place_id") REFERENCES "Place" ("id") ON DELETE CASCADE;

-- +goose Down
DROP TABLE IF EXISTS "UserFavoritePlace" CASCADE;
