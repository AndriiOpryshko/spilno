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
INSERT INTO board_member_revisions (id, association_id, board_member_id, full_name, sign_in_email, term_start, term_end, position, decision_id, created_by_member_id)
VALUES (sqlc.arg(id), sqlc.arg(association_id), sqlc.arg(board_member_id), sqlc.arg(full_name), sqlc.narg(sign_in_email),
        sqlc.arg(term_start), sqlc.narg(term_end), sqlc.arg(position), sqlc.narg(decision_id), sqlc.narg(created_by_member_id));

-- name: InsertBoardMember :exec
INSERT INTO board_members (id, association_id) VALUES (sqlc.arg(id), sqlc.arg(association_id));

-- name: LatestRevisionOfMember :one
SELECT full_name::text AS full_name, COALESCE(sign_in_email, '')::text AS sign_in_email, term_start, term_end,
       COALESCE(position, 'member')::text AS position, COALESCE(decision_id::text, '')::text AS decision_id
FROM board_member_revisions
WHERE association_id = sqlc.arg(association_id) AND board_member_id = sqlc.arg(board_member_id)
ORDER BY created_at DESC, id DESC
LIMIT 1;

-- name: CurrentMemberIDByEmail :one
SELECT board_member_id
FROM (
  SELECT DISTINCT ON (board_member_id) board_member_id, sign_in_email, term_start, term_end
  FROM board_member_revisions
  WHERE association_id = sqlc.arg(association_id)
  ORDER BY board_member_id, created_at DESC, id DESC
) latest
WHERE lower(sign_in_email) = lower(sqlc.arg(email)::text)
  AND term_start <= sqlc.arg(on_date)::date
  AND (term_end IS NULL OR term_end >= sqlc.arg(on_date)::date)
LIMIT 1;

-- name: ListRosterEntries :many
-- One row per entry (its latest revision), newest entry first by the (created_at, id) keyset; a NULL
-- on_date, after_created_at or entry_id leaves that filter off.
SELECT bm.id AS id, bm.created_at AS entry_created_at,
       r.full_name::text AS full_name, r.sign_in_email, r.term_start, r.term_end,
       COALESCE(r.position, 'member')::text AS position, r.decision_id,
       r.created_by_member_id AS revised_by_member_id, r.created_at AS revised_at
FROM board_members bm
JOIN LATERAL (
  SELECT * FROM board_member_revisions
  WHERE board_member_id = bm.id AND association_id = bm.association_id
  ORDER BY created_at DESC, id DESC
  LIMIT 1
) r ON true
WHERE bm.association_id = sqlc.arg(association_id)
  AND (sqlc.narg(entry_id)::uuid IS NULL OR bm.id = sqlc.narg(entry_id)::uuid)
  AND (sqlc.narg(on_date)::date IS NULL
   OR (r.term_start <= sqlc.narg(on_date)::date AND (r.term_end IS NULL OR r.term_end >= sqlc.narg(on_date)::date)))
  AND (sqlc.narg(after_created_at)::timestamptz IS NULL
   OR (bm.created_at, bm.id) < (sqlc.narg(after_created_at)::timestamptz, sqlc.narg(after_id)::uuid))
ORDER BY bm.created_at DESC, bm.id DESC
LIMIT sqlc.arg(row_limit);

-- name: ActorEmail :one
-- The sign-in address db.InAssociation set for this transaction; empty when none.
SELECT COALESCE(current_setting('app.actor_email', true), '')::text AS email;
