-- Phase 1 of 3 (expand) on the existing meetings table: association_id arrives nullable and is
-- written by CreateMeeting from this release on (dual-write in app code, not a trigger).
-- chair_board_member_id stays nullable: existing meetings have no recorded головуючий (AC-28).
ALTER TABLE meetings ADD COLUMN IF NOT EXISTS association_id UUID;
ALTER TABLE meetings ADD COLUMN IF NOT EXISTS chair_board_member_id UUID REFERENCES board_members (id);
