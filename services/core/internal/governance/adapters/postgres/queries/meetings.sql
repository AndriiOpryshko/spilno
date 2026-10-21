-- name: CreateMeeting :one
INSERT INTO meetings (title, kind, held_at, ends_at, location, quorum_required)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetMeeting :one
SELECT * FROM meetings WHERE id = $1;

-- name: ListMeetings :many
-- Pages newest first by the (created_at, id) keyset; a NULL after_created_at
-- starts from the top.
SELECT * FROM meetings
WHERE sqlc.narg(after_created_at)::timestamptz IS NULL
   OR (created_at, id) < (sqlc.narg(after_created_at)::timestamptz, sqlc.narg(after_id)::uuid)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(row_limit);
