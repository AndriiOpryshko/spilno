-- +goose Up
-- Promoted from docs/features/board-protocol/migrations/20261027120021_create_protocol_immutability_triggers.up.sql.
-- DOCUMENTED EXCEPTION to "no CREATE TRIGGER" (.claude/rules/migrations.md): board-protocol ADR-0007,
-- QG-1 and QG-9. protocol_snapshots is append-only; no protocol is ever deleted, in any state.
-- Row triggers do not fire on TRUNCATE, so each table also gets a statement-level TRUNCATE guard.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION reject_protocol_record_change() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION '% on % is not allowed: protocol records are immutable (ADR-0007)', TG_OP, TG_TABLE_NAME
        USING ERRCODE = 'restrict_violation';
END;
$$;
-- +goose StatementEnd

CREATE OR REPLACE TRIGGER protocol_snapshots_append_only
    BEFORE UPDATE OR DELETE ON protocol_snapshots
    FOR EACH ROW EXECUTE FUNCTION reject_protocol_record_change();

CREATE OR REPLACE TRIGGER protocol_snapshots_no_truncate
    BEFORE TRUNCATE ON protocol_snapshots
    FOR EACH STATEMENT EXECUTE FUNCTION reject_protocol_record_change();

CREATE OR REPLACE TRIGGER protocols_no_delete
    BEFORE DELETE ON protocols
    FOR EACH ROW EXECUTE FUNCTION reject_protocol_record_change();

CREATE OR REPLACE TRIGGER protocols_no_truncate
    BEFORE TRUNCATE ON protocols
    FOR EACH STATEMENT EXECUTE FUNCTION reject_protocol_record_change();

-- +goose Down
DROP TRIGGER IF EXISTS protocols_no_truncate ON protocols;
DROP TRIGGER IF EXISTS protocols_no_delete ON protocols;
DROP TRIGGER IF EXISTS protocol_snapshots_no_truncate ON protocol_snapshots;
DROP TRIGGER IF EXISTS protocol_snapshots_append_only ON protocol_snapshots;
DROP FUNCTION IF EXISTS reject_protocol_record_change();
