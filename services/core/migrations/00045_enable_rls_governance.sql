-- Staged as docs/features/association-onboarding/migrations/NNNNN_enable_rls_governance.sql (order in MANIFEST.md).
-- Promote with `mise run migrate:new enable_rls_governance`; never hand-type the number.
-- Owner: governance. Origin: SAD ADR-0003; PRD §6 Isolation.
-- Row-level security (ADR-0003): rows are visible to spilno_app only for the association set by
-- platform/db.InAssociation (set_config('app.association_id', ..., true)); unset -> no rows (fail
-- closed). FORCE applies it to the table owner too; the migration/owner role therefore needs
-- BYPASSRLS (a superuser in compose and testdb has it) so goose and backfills still see every row.

-- +goose Up
ALTER TABLE meetings ENABLE ROW LEVEL SECURITY;
ALTER TABLE meetings FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON meetings
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE board_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE board_members FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON board_members
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE board_member_revisions ENABLE ROW LEVEL SECURITY;
ALTER TABLE board_member_revisions FORCE ROW LEVEL SECURITY;

-- A person also sees their own roster entries in every association, so governance.MembershipsOf
-- works without a bypass (ADR-0003). Writes stay limited to the association set.
CREATE POLICY association_isolation ON board_member_revisions
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid
        OR lower(sign_in_email) = lower(NULLIF(current_setting('app.actor_email', true), '')))
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE association_settings ENABLE ROW LEVEL SECURITY;
ALTER TABLE association_settings FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON association_settings
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE protocol_registers ENABLE ROW LEVEL SECURITY;
ALTER TABLE protocol_registers FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON protocol_registers
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE files ENABLE ROW LEVEL SECURITY;
ALTER TABLE files FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON files
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE cases ENABLE ROW LEVEL SECURITY;
ALTER TABLE cases FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON cases
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE case_documents ENABLE ROW LEVEL SECURITY;
ALTER TABLE case_documents FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON case_documents
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE protocols ENABLE ROW LEVEL SECURITY;
ALTER TABLE protocols FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON protocols
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE protocol_transitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE protocol_transitions FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON protocol_transitions
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE agenda_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE agenda_items FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON agenda_items
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE agenda_item_versions ENABLE ROW LEVEL SECURITY;
ALTER TABLE agenda_item_versions FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON agenda_item_versions
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE assistant_turns ENABLE ROW LEVEL SECURITY;
ALTER TABLE assistant_turns FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON assistant_turns
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE assistant_turn_documents ENABLE ROW LEVEL SECURITY;
ALTER TABLE assistant_turn_documents FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON assistant_turn_documents
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE decision_slots ENABLE ROW LEVEL SECURITY;
ALTER TABLE decision_slots FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON decision_slots
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE slot_values ENABLE ROW LEVEL SECURITY;
ALTER TABLE slot_values FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON slot_values
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE slot_confirmations ENABLE ROW LEVEL SECURITY;
ALTER TABLE slot_confirmations FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON slot_confirmations
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE poll_rounds ENABLE ROW LEVEL SECURITY;
ALTER TABLE poll_rounds FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON poll_rounds
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE poll_votes ENABLE ROW LEVEL SECURITY;
ALTER TABLE poll_votes FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON poll_votes
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE print_copies ENABLE ROW LEVEL SECURITY;
ALTER TABLE print_copies FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON print_copies
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE meeting_facts ENABLE ROW LEVEL SECURITY;
ALTER TABLE meeting_facts FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON meeting_facts
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE meeting_attendees ENABLE ROW LEVEL SECURITY;
ALTER TABLE meeting_attendees FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON meeting_attendees
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE meeting_item_votes ENABLE ROW LEVEL SECURITY;
ALTER TABLE meeting_item_votes FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON meeting_item_votes
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE protocol_snapshots ENABLE ROW LEVEL SECURITY;
ALTER TABLE protocol_snapshots FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON protocol_snapshots
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE registry_entries ENABLE ROW LEVEL SECURITY;
ALTER TABLE registry_entries FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON registry_entries
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE integrity_runs ENABLE ROW LEVEL SECURITY;
ALTER TABLE integrity_runs FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON integrity_runs
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

ALTER TABLE integrity_findings ENABLE ROW LEVEL SECURITY;
ALTER TABLE integrity_findings FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON integrity_findings
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

-- +goose Down
DROP POLICY IF EXISTS association_isolation ON integrity_findings;
ALTER TABLE integrity_findings NO FORCE ROW LEVEL SECURITY;
ALTER TABLE integrity_findings DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON integrity_runs;
ALTER TABLE integrity_runs NO FORCE ROW LEVEL SECURITY;
ALTER TABLE integrity_runs DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON registry_entries;
ALTER TABLE registry_entries NO FORCE ROW LEVEL SECURITY;
ALTER TABLE registry_entries DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON protocol_snapshots;
ALTER TABLE protocol_snapshots NO FORCE ROW LEVEL SECURITY;
ALTER TABLE protocol_snapshots DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON meeting_item_votes;
ALTER TABLE meeting_item_votes NO FORCE ROW LEVEL SECURITY;
ALTER TABLE meeting_item_votes DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON meeting_attendees;
ALTER TABLE meeting_attendees NO FORCE ROW LEVEL SECURITY;
ALTER TABLE meeting_attendees DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON meeting_facts;
ALTER TABLE meeting_facts NO FORCE ROW LEVEL SECURITY;
ALTER TABLE meeting_facts DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON print_copies;
ALTER TABLE print_copies NO FORCE ROW LEVEL SECURITY;
ALTER TABLE print_copies DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON poll_votes;
ALTER TABLE poll_votes NO FORCE ROW LEVEL SECURITY;
ALTER TABLE poll_votes DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON poll_rounds;
ALTER TABLE poll_rounds NO FORCE ROW LEVEL SECURITY;
ALTER TABLE poll_rounds DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON slot_confirmations;
ALTER TABLE slot_confirmations NO FORCE ROW LEVEL SECURITY;
ALTER TABLE slot_confirmations DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON slot_values;
ALTER TABLE slot_values NO FORCE ROW LEVEL SECURITY;
ALTER TABLE slot_values DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON decision_slots;
ALTER TABLE decision_slots NO FORCE ROW LEVEL SECURITY;
ALTER TABLE decision_slots DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON assistant_turn_documents;
ALTER TABLE assistant_turn_documents NO FORCE ROW LEVEL SECURITY;
ALTER TABLE assistant_turn_documents DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON assistant_turns;
ALTER TABLE assistant_turns NO FORCE ROW LEVEL SECURITY;
ALTER TABLE assistant_turns DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON agenda_item_versions;
ALTER TABLE agenda_item_versions NO FORCE ROW LEVEL SECURITY;
ALTER TABLE agenda_item_versions DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON agenda_items;
ALTER TABLE agenda_items NO FORCE ROW LEVEL SECURITY;
ALTER TABLE agenda_items DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON protocol_transitions;
ALTER TABLE protocol_transitions NO FORCE ROW LEVEL SECURITY;
ALTER TABLE protocol_transitions DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON protocols;
ALTER TABLE protocols NO FORCE ROW LEVEL SECURITY;
ALTER TABLE protocols DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON case_documents;
ALTER TABLE case_documents NO FORCE ROW LEVEL SECURITY;
ALTER TABLE case_documents DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON cases;
ALTER TABLE cases NO FORCE ROW LEVEL SECURITY;
ALTER TABLE cases DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON files;
ALTER TABLE files NO FORCE ROW LEVEL SECURITY;
ALTER TABLE files DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON protocol_registers;
ALTER TABLE protocol_registers NO FORCE ROW LEVEL SECURITY;
ALTER TABLE protocol_registers DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON association_settings;
ALTER TABLE association_settings NO FORCE ROW LEVEL SECURITY;
ALTER TABLE association_settings DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON board_member_revisions;
ALTER TABLE board_member_revisions NO FORCE ROW LEVEL SECURITY;
ALTER TABLE board_member_revisions DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON board_members;
ALTER TABLE board_members NO FORCE ROW LEVEL SECURITY;
ALTER TABLE board_members DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS association_isolation ON meetings;
ALTER TABLE meetings NO FORCE ROW LEVEL SECURITY;
ALTER TABLE meetings DISABLE ROW LEVEL SECURITY;
