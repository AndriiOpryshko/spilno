-- name: GetFileForDownload :one
-- One stored file of the association and whether a signed snapshot pins it (its id is a string value
-- anywhere inside the snapshot document); another association's id finds nothing.
SELECT f.drive_file_id, f.sha256, f.size_bytes, f.media_type, f.original_name,
       EXISTS (
           SELECT 1 FROM protocol_snapshots s
           WHERE s.association_id = f.association_id
             AND s.kind = 'signed'
             AND jsonb_path_exists(s.snapshot, '$.** ? (@ == $fid)', jsonb_build_object('fid', f.id::text))
       )::bool AS pinned
FROM files f
WHERE f.association_id = sqlc.arg(association_id)
  AND f.id = sqlc.arg(id);
