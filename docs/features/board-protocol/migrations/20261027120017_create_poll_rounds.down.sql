DROP INDEX IF EXISTS idx_poll_votes_board_member;
DROP INDEX IF EXISTS idx_poll_votes_agenda_item;
DROP INDEX IF EXISTS idx_poll_votes_round_item_member_created;
DROP TABLE IF EXISTS poll_votes;
DROP INDEX IF EXISTS idx_poll_rounds_closed_by;
DROP INDEX IF EXISTS idx_poll_rounds_opened_by;
DROP TABLE IF EXISTS poll_rounds;
