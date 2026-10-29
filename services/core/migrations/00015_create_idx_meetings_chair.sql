-- +goose NO TRANSACTION
-- +goose Up
-- Promoted from docs/features/board-protocol/migrations/20261027120011_create_idx_meetings_chair.up.sql.
-- Existing table: CONCURRENTLY, one statement, outside a transaction (goose: NO TRANSACTION).
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_meetings_chair_board_member_id
    ON meetings (chair_board_member_id);

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS idx_meetings_chair_board_member_id;
