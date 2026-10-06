-- +goose Up
ALTER TABLE "User" ADD COLUMN email_verified BOOLEAN DEFAULT false;
ALTER TABLE "User" ADD COLUMN phone_verified BOOLEAN DEFAULT false;
ALTER TABLE "User" ADD COLUMN two_factor_enabled BOOLEAN DEFAULT false;
ALTER TABLE "User" ADD COLUMN two_factor_secret TEXT;
ALTER TABLE "User" ADD COLUMN backup_codes TEXT[];
ALTER TABLE "Place" ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- +goose Down
ALTER TABLE "User" DROP COLUMN email_verified;
ALTER TABLE "User" DROP COLUMN phone_verified;
ALTER TABLE "User" DROP COLUMN two_factor_enabled;
ALTER TABLE "User" DROP COLUMN two_factor_secret;
ALTER TABLE "User" DROP COLUMN backup_codes;
ALTER TABLE "Place" DROP COLUMN created_at;
