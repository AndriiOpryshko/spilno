-- Owner: finance. Origin: PRD AC-01, AC-09, §9; ADR-0002.
-- Expand only: both columns stay nullable for good (NULL = no equivalent: a UAH contract, or the rate
-- could not be read yet). Set together, once, by the app; no CHECK (business invariant).

-- +goose Up
ALTER TABLE contracts ADD COLUMN IF NOT EXISTS uah_amount_kopiyky BIGINT;
ALTER TABLE contracts ADD COLUMN IF NOT EXISTS uah_rate_micros BIGINT;

-- +goose Down
ALTER TABLE contracts DROP COLUMN IF EXISTS uah_rate_micros;
ALTER TABLE contracts DROP COLUMN IF EXISTS uah_amount_kopiyky;
