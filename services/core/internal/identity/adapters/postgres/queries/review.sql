-- name: ListPendingApplications :many
SELECT a.id, a.kind, a.edrpou, a.association_id, a.state, a.created_at,
       f.association_name, f.created_at AS filed_at, d.reason AS rejection_reason
FROM applications a
JOIN LATERAL (
    SELECT association_name, created_at FROM application_filings
    WHERE application_id = a.id ORDER BY created_at DESC, id DESC LIMIT 1
) f ON true
LEFT JOIN LATERAL (
    SELECT reason FROM application_decisions
    WHERE application_id = a.id AND outcome = 'rejected' ORDER BY created_at DESC, id DESC LIMIT 1
) d ON true
WHERE a.state = @state
  AND (sqlc.narg(cursor_created_at)::timestamptz IS NULL
       OR (a.created_at, a.id) > (sqlc.narg(cursor_created_at)::timestamptz, sqlc.narg(cursor_id)::uuid))
ORDER BY a.created_at, a.id
LIMIT @page_size;

-- name: ListDecidedApplications :many
SELECT a.id, a.kind, a.edrpou, a.association_id, a.state, a.created_at,
       f.association_name, f.created_at AS filed_at, d.reason AS rejection_reason
FROM applications a
JOIN LATERAL (
    SELECT association_name, created_at FROM application_filings
    WHERE application_id = a.id ORDER BY created_at DESC, id DESC LIMIT 1
) f ON true
LEFT JOIN LATERAL (
    SELECT reason FROM application_decisions
    WHERE application_id = a.id AND outcome = 'rejected' ORDER BY created_at DESC, id DESC LIMIT 1
) d ON true
WHERE a.state = @state
  AND (sqlc.narg(cursor_created_at)::timestamptz IS NULL
       OR (a.created_at, a.id) < (sqlc.narg(cursor_created_at)::timestamptz, sqlc.narg(cursor_id)::uuid))
ORDER BY a.created_at DESC, a.id DESC
LIMIT @page_size;

-- name: GetApplication :one
SELECT a.id, a.kind, a.edrpou, a.association_id, a.state, a.created_at, c.email AS applicant_email,
       f.association_name, f.created_at AS filed_at, d.reason AS rejection_reason
FROM applications a
JOIN accounts c ON c.id = a.account_id
JOIN LATERAL (
    SELECT association_name, created_at FROM application_filings
    WHERE application_id = a.id ORDER BY created_at DESC, id DESC LIMIT 1
) f ON true
LEFT JOIN LATERAL (
    SELECT reason FROM application_decisions
    WHERE application_id = a.id AND outcome = 'rejected' ORDER BY created_at DESC, id DESC LIMIT 1
) d ON true
WHERE a.id = @id;

-- name: ListFilingsOfApplication :many
SELECT f.id, f.association_name, f.association_address, f.applicant_full_name, f.position,
       f.term_start, f.term_end, f.created_at,
       e.id AS extract_id, e.media_type, e.original_name, e.sha256, e.size_bytes
FROM application_filings f
JOIN application_extracts e ON e.id = f.extract_id
WHERE f.application_id = @application_id
ORDER BY f.created_at DESC, f.id DESC;

-- name: ListChecksOfApplication :many
SELECT k.id, k.filing_id, k.signature_valid, k.signer_full_name, k.signed_on, k.head_full_name,
       k.extract_edrpou, k.extract_dated_on, k.created_at
FROM application_checks k
JOIN application_filings f ON f.id = k.filing_id
WHERE f.application_id = @application_id
ORDER BY k.created_at, k.id;

-- name: ListDecisionsOfApplication :many
SELECT id, filing_id, association_id, outcome, reason, relied_on_document, created_at
FROM application_decisions
WHERE application_id = @application_id
ORDER BY created_at DESC, id DESC;

-- name: ListCompetingApplications :many
SELECT a.id FROM applications a
WHERE a.state = 'pending' AND a.id <> @id
  AND a.edrpou = (SELECT edrpou FROM applications WHERE id = @id)
ORDER BY a.created_at, a.id;

-- name: CreateCheck :one
INSERT INTO application_checks (
    id, filing_id, signature_valid, signer_full_name, signed_on, head_full_name,
    extract_edrpou, extract_dated_on, checked_by_email
) VALUES (
    @id, @filing_id, @signature_valid, @signer_full_name, @signed_on, @head_full_name,
    @extract_edrpou, @extract_dated_on, @checked_by_email
)
RETURNING created_at;

-- name: GetLatestCheck :one
SELECT id, filing_id, signature_valid, signer_full_name, signed_on, head_full_name,
       extract_edrpou, extract_dated_on, created_at
FROM application_checks
WHERE filing_id = @filing_id
ORDER BY created_at DESC, id DESC
LIMIT 1;

-- name: GetExtractOfFiling :one
SELECT e.content, e.media_type, e.original_name
FROM application_filings f
JOIN application_extracts e ON e.id = f.extract_id
WHERE f.id = @filing_id AND f.application_id = @application_id;

-- name: FilingBelongsToApplication :one
SELECT id FROM application_filings WHERE id = @filing_id AND application_id = @application_id;
