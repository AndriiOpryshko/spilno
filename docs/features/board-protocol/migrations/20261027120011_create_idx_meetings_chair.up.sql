-- Existing table: CONCURRENTLY, one statement, outside a transaction (goose: NO TRANSACTION).
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_meetings_chair_board_member_id
    ON meetings (chair_board_member_id);
