-- name: ListAssociations :many
SELECT id, name, edrpou, state, confirmed_by_decision_id, created_at
FROM associations
WHERE sqlc.narg(cursor_created_at)::timestamptz IS NULL
   OR (created_at, id) < (sqlc.narg(cursor_created_at)::timestamptz, sqlc.narg(cursor_id)::uuid)
ORDER BY created_at DESC, id DESC
LIMIT @page_size;

-- name: GetAssociationStateForUpdate :one
SELECT state FROM associations WHERE id = @id FOR UPDATE;

-- name: UpdateAssociationState :exec
UPDATE associations SET state = @state WHERE id = @id;

-- name: CreateStateChange :one
INSERT INTO association_state_changes (id, association_id, to_state, reason, changed_by_email)
VALUES (@id, @association_id, @to_state, @reason, @changed_by_email)
RETURNING created_at;
