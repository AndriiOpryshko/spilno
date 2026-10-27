-- +goose Up
-- Phase 1 of 3 (expand): association_id arrives nullable; CreateMeeting writes it from this
-- release on (dual-write in app code, not a trigger). No FK: associations belong to identity,
-- meetings to governance (repo ADR 0002).
ALTER TABLE meetings ADD COLUMN association_id uuid;

-- +goose Down
ALTER TABLE meetings DROP COLUMN association_id;
