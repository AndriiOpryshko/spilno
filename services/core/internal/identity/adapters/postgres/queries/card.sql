-- name: BriefAssociations :many
SELECT id, name, state FROM associations WHERE id = ANY(@ids::uuid[]);

-- name: GetCardHead :one
SELECT name, state, edrpou, address, confirmed_by_decision_id FROM associations WHERE id = @id;

-- name: GetCardHeadFullName :one
SELECT full_name::text AS full_name FROM (
    SELECT DISTINCT ON (board_member_id) full_name, position, term_start, term_end
    FROM board_member_revisions
    WHERE association_id = @association_id
    ORDER BY board_member_id, created_at DESC, id DESC
) r
WHERE position = 'head' AND term_start <= @today::date AND (term_end IS NULL OR term_end >= @today::date)
ORDER BY term_start DESC LIMIT 1;

-- name: ListCardHistory :many
SELECT CASE WHEN a.kind = 'connection' THEN 'connected' ELSE 'transfer_approved' END::text AS kind,
       d.created_at AS occurred_at, d.id AS decision_id, f.created_at AS filed_at, NULL::text AS reason,
       k.head_full_name,
       COALESCE(ended.full_name, '')::text AS ended_head_full_name,
       k.signer_full_name, f.extract_id, k.extract_dated_on, d.id AS tie
FROM application_decisions d
JOIN applications a ON a.id = d.application_id
JOIN application_filings f ON f.id = d.filing_id
LEFT JOIN application_checks k ON k.id = d.check_id
-- A decision writes the ended entry first and the seated one after it (approval.go), so the entry it ended
-- is the revision of the decision that a later revision of the same decision, of another board member, follows.
LEFT JOIN (
    SELECT DISTINCT ON (r.decision_id) r.decision_id, r.full_name
    FROM board_member_revisions r
    WHERE r.association_id = @association_id AND r.decision_id IS NOT NULL
      AND EXISTS (
        SELECT 1 FROM board_member_revisions s
        WHERE s.association_id = r.association_id AND s.decision_id = r.decision_id
          AND s.board_member_id <> r.board_member_id AND s.id > r.id)
    ORDER BY r.decision_id, r.id DESC
) ended ON ended.decision_id = d.id
WHERE d.association_id = @association_id AND d.outcome = 'approved'
UNION ALL
SELECT CASE WHEN c.to_state = 'disabled' THEN 'disabled' ELSE 'enabled' END::text,
       c.created_at, NULL::uuid, NULL::timestamptz, c.reason, NULL::text, ''::text, NULL::text, NULL::uuid, NULL::date, c.id
FROM association_state_changes c
WHERE c.association_id = @association_id
ORDER BY occurred_at DESC, tie DESC;

-- name: GetCardExtract :one
SELECT x.original_name, x.media_type, x.content, x.sha256
FROM application_decisions d
JOIN application_filings f ON f.id = d.filing_id
JOIN application_extracts x ON x.id = f.extract_id
WHERE d.association_id = @association_id AND d.outcome = 'approved' AND x.id = @extract_id
LIMIT 1;
