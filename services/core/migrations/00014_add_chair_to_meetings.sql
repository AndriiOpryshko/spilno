-- +goose Up
-- Promoted from docs/features/board-protocol/migrations/20261027120007_add_association_and_chair_to_meetings.up.sql.
-- Only the chair column: meetings.association_id is already live (00005-00008).
-- chair_board_member_id stays nullable: existing meetings have no recorded головуючий (AC-28).
ALTER TABLE meetings ADD COLUMN IF NOT EXISTS chair_board_member_id UUID REFERENCES board_members (id);

-- +goose Down
ALTER TABLE meetings DROP COLUMN IF EXISTS chair_board_member_id;
