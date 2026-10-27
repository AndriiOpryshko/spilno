-- +goose Up
-- Phase 3 of 3 (contract): only after CreateMeeting writes association_id in the running release.
ALTER TABLE meetings ALTER COLUMN association_id SET NOT NULL;

-- +goose Down
ALTER TABLE meetings ALTER COLUMN association_id DROP NOT NULL;
