DROP TRIGGER IF EXISTS protocols_no_truncate ON protocols;
DROP TRIGGER IF EXISTS protocols_no_delete ON protocols;
DROP TRIGGER IF EXISTS protocol_snapshots_no_truncate ON protocol_snapshots;
DROP TRIGGER IF EXISTS protocol_snapshots_append_only ON protocol_snapshots;
DROP FUNCTION IF EXISTS reject_protocol_record_change();
