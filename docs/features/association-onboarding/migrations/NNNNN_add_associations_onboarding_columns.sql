-- Staged as docs/features/association-onboarding/migrations/NNNNN_add_associations_onboarding_columns.sql (order in MANIFEST.md).
-- Promote with `mise run migrate:new add_associations_onboarding_columns`; never hand-type the number.
-- Owner: identity. Origin: PRD §9, AC-04, AC-29; SAD ADR-0007.
-- Phase 1 of 3 (expand) for state. PR 1. edrpou, address, confirmed_by_decision_id stay nullable for good.
-- edrpou is NULL for the association set up by hand until AC-19c; address is set by the board.
-- confirmed_by_decision_id NULL = «підключено вручну, без витягу» (AC-29).
-- App code in this PR writes state on every new association; reads treat NULL as connected.

-- +goose Up
ALTER TABLE associations ADD COLUMN IF NOT EXISTS edrpou VARCHAR(8);
ALTER TABLE associations ADD COLUMN IF NOT EXISTS address VARCHAR(500);
ALTER TABLE associations ADD COLUMN IF NOT EXISTS state VARCHAR(32);
ALTER TABLE associations ADD COLUMN IF NOT EXISTS confirmed_by_decision_id UUID REFERENCES application_decisions (id);

-- +goose Down
ALTER TABLE associations DROP COLUMN IF EXISTS confirmed_by_decision_id;
ALTER TABLE associations DROP COLUMN IF EXISTS state;
ALTER TABLE associations DROP COLUMN IF EXISTS address;
ALTER TABLE associations DROP COLUMN IF EXISTS edrpou;
