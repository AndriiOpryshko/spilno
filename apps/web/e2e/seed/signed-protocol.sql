-- A signed protocol for the e2e suite, written over the owner connection: nothing in the app can lock a
-- protocol before board-protocol T29, so the реєстр and the signed-copy download (AC-26) get their data here.
--
-- Run by seedSignedProtocol (association-support.ts) through psql with the variables
--   association, member, case_id, scan_file, print_file, number, meeting_date, subject.
-- The two files are rows the member already uploaded as documents of the справа, so their bytes sit in
-- the API's in-memory storage (SPILNO_DEV_STORAGE) and the download serves them.
--
-- The snapshot has the shape of governancetest.SeedSigned (schema_version 1), without the working rows
-- the screens do not read (print_copies, meeting_facts, poll rounds). Its fingerprint is the digest of the
-- stored text, not of the canonical form: nothing in the e2e flows verifies it.
\set ON_ERROR_STOP on

-- protocols and protocol_snapshots cannot be deleted (ADR-0007), so this must never touch a database
-- that is not the throwaway e2e one. That database says so itself:
--   ALTER DATABASE <name> SET spilno.e2e = 'on';      (undo: ALTER DATABASE <name> RESET spilno.e2e;)
-- Only the setting stored on the database counts: one that comes from the role, the connection
-- options or the session would mark every database that role can reach.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_db_role_setting s
        JOIN pg_database d ON d.oid = s.setdatabase
        WHERE d.datname = current_database() AND s.setrole = 0 AND 'spilno.e2e=on' = ANY (s.setconfig)
    ) THEN
        RAISE EXCEPTION 'refusing to seed: database % is not marked as the e2e one (ALTER DATABASE ... SET spilno.e2e)', current_database();
    END IF;
END
$$;

BEGIN;

SELECT uuidv7() AS protocol, uuidv7() AS meeting, uuidv7() AS item, uuidv7() AS snapshot, uuidv7() AS print_copy \gset

INSERT INTO meetings (id, title, held_at, association_id, chair_board_member_id)
VALUES (:'meeting', 'Test meeting', :'meeting_date'::date + interval '16 hours', :'association', :'member');

INSERT INTO protocols (id, association_id, meeting_id, state, revision, number, created_by_member_id)
VALUES (:'protocol', :'association', :'meeting', 'signed', 9, :number, :'member');

INSERT INTO agenda_items (id, association_id, protocol_id, case_id, position, created_by_member_id)
VALUES (:'item', :'association', :'protocol', :'case_id', 1, :'member');

-- No file of this association with the given id leaves the snapshot unwritten, and the registry entry
-- below then fails on its foreign key.
WITH file_refs AS (
    SELECT f.id,
           jsonb_build_object(
               'id', f.id::text, 'original_name', f.original_name, 'media_type', f.media_type,
               'size_bytes', f.size_bytes, 'sha256', encode(f.sha256, 'hex'),
               'uploaded_by_member_id', f.uploaded_by_member_id::text,
               'created_at', to_char(f.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
               'drive_file_id', f.drive_file_id, 'drive_revision_id', f.drive_revision_id
           ) AS ref
    FROM files f
    WHERE f.association_id = :'association'
), stamp AS (
    SELECT to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS at
), doc AS (
    SELECT jsonb_build_object(
        'meeting', jsonb_build_object(
            'id', :'meeting', 'held_at', :'meeting_date' || 'T16:00:00Z', 'meeting_date', :'meeting_date',
            'chair_board_member_id', :'member'),
        'agenda_items', jsonb_build_array(jsonb_build_object(
            'id', :'item', 'position', 1, 'case_id', :'case_id', 'subject', :'subject',
            'fragments', jsonb_build_array(jsonb_build_object('text', :'subject')),
            'wording_author_member_id', :'member', 'wording_revised_at', stamp.at, 'slots', jsonb_build_array())),
        'documents', jsonb_build_array(),
        'print_copy', jsonb_build_object(
            'id', :'print_copy', 'protocol_revision', 8, 'number', :number, 'file', print.ref,
            'requested_by_member_id', :'member', 'created_at', stamp.at, 'voided_at', NULL),
        'meeting_facts', jsonb_build_object(
            'signed_scan', scan.ref,
            'present_member_ids', jsonb_build_array(:'member'),
            'item_votes', jsonb_build_array(jsonb_build_object(
                'agenda_item_id', :'item', 'votes_for', 1, 'votes_against', 0, 'votes_abstained', 0)),
            'has_handwritten_corrections', false, 'corrections_note', NULL,
            'entered_by_member_id', :'member', 'entered_at', stamp.at),
        'poll_history', jsonb_build_array(),
        'locked_by_member_id', :'member',
        'locked_at', stamp.at
    ) AS body
    FROM stamp, file_refs scan, file_refs print
    WHERE scan.id = :'scan_file' AND print.id = :'print_file'
)
INSERT INTO protocol_snapshots
    (id, association_id, protocol_id, kind, number, meeting_date, schema_version, snapshot, fingerprint, created_by_member_id)
SELECT :'snapshot', :'association', :'protocol', 'signed', :number, :'meeting_date', 1, doc.body,
       sha256(convert_to(doc.body::text, 'UTF8')), :'member'
FROM doc;

INSERT INTO registry_entries (id, association_id, snapshot_id, agenda_item_id, number, meeting_date, subject)
VALUES (uuidv7(), :'association', :'snapshot', :'item', :number, :'meeting_date', :'subject');

COMMIT;
