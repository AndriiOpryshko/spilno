-- +goose Up
-- Phase 2 of 3 (backfill): every existing meeting belongs to the one bootstrap association
-- (ADR-0015). See docs/features/board-protocol/migrations/backfill-meetings-association-id.md.
UPDATE meetings
SET association_id = '00000000-0000-7000-8000-000000000001'
WHERE association_id IS NULL;

-- +goose Down
UPDATE meetings
SET association_id = NULL
WHERE association_id = '00000000-0000-7000-8000-000000000001';
