-- name: CurrentMembership :one
-- The member's latest revision decides; a term is current from term_start
-- through term_end inclusive, an open end is current.
SELECT position::text AS position
FROM (
  SELECT DISTINCT ON (board_member_id) sign_in_email, term_start, term_end, COALESCE(position, 'member')::text AS position
  FROM board_member_revisions
  WHERE association_id = sqlc.arg(association_id)
  ORDER BY board_member_id, created_at DESC, id DESC
) latest
WHERE lower(sign_in_email) = lower(sqlc.arg(email)::text)
  AND term_start <= sqlc.arg(on_date)::date
  AND (term_end IS NULL OR term_end >= sqlc.arg(on_date)::date)
LIMIT 1;

-- name: MembershipsOf :many
SELECT association_id, position::text AS position
FROM (
  SELECT DISTINCT ON (board_member_id) association_id, sign_in_email, term_start, term_end, COALESCE(position, 'member')::text AS position
  FROM board_member_revisions
  ORDER BY board_member_id, created_at DESC, id DESC
) latest
WHERE lower(sign_in_email) = lower(sqlc.arg(email)::text)
  AND term_start <= sqlc.arg(on_date)::date
  AND (term_end IS NULL OR term_end >= sqlc.arg(on_date)::date)
ORDER BY association_id;

-- name: MemberSignInAddresses :many
SELECT sign_in_email::text AS sign_in_email
FROM (
  SELECT DISTINCT ON (board_member_id) sign_in_email, term_start, term_end
  FROM board_member_revisions
  WHERE association_id = sqlc.arg(association_id)
  ORDER BY board_member_id, created_at DESC, id DESC
) latest
WHERE sign_in_email IS NOT NULL
  AND term_start <= sqlc.arg(on_date)::date
  AND (term_end IS NULL OR term_end >= sqlc.arg(on_date)::date)
ORDER BY sign_in_email;

-- name: InsertBoardMemberRevision :exec
INSERT INTO board_member_revisions (id, association_id, board_member_id, full_name, sign_in_email, term_start, term_end, position, decision_id)
VALUES (sqlc.arg(id), sqlc.arg(association_id), sqlc.arg(board_member_id), sqlc.arg(full_name), sqlc.narg(sign_in_email),
        sqlc.arg(term_start), sqlc.narg(term_end), sqlc.arg(position), sqlc.narg(decision_id));
