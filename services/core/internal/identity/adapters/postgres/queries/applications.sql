-- name: CreateExtract :exec
INSERT INTO application_extracts (id, account_id, content, sha256, size_bytes, media_type, original_name)
VALUES (@id, @account_id, @content, @sha256, @size_bytes, @media_type, @original_name);

-- name: CreateApplication :one
INSERT INTO applications (id, kind, account_id, edrpou, association_id, state)
VALUES (@id, @kind, @account_id, @edrpou, @association_id, @state)
RETURNING created_at;

-- name: CreateFiling :one
INSERT INTO application_filings (
    id, application_id, extract_id, association_name, association_address,
    applicant_full_name, position, term_start, term_end
) VALUES (
    @id, @application_id, @extract_id, @association_name, @association_address,
    @applicant_full_name, @position, @term_start, @term_end
)
RETURNING created_at;

-- name: GetAssociationByEdrpou :one
SELECT id FROM associations WHERE edrpou = @edrpou;

-- name: FindOpenApplicationOfAccount :one
SELECT id FROM applications WHERE account_id = @account_id AND edrpou = @edrpou AND state = 'pending';

-- name: ListApplicationsOfAccount :many
SELECT a.id, a.kind, a.edrpou, a.association_id, a.state, a.created_at,
       f.association_name, f.created_at AS filed_at,
       (SELECT d.reason FROM application_decisions d
         WHERE d.application_id = a.id AND d.outcome = 'rejected'
         ORDER BY d.created_at DESC, d.id DESC LIMIT 1) AS rejection_reason
FROM applications a
JOIN LATERAL (SELECT x.association_name, x.created_at FROM application_filings x
               WHERE x.application_id = a.id ORDER BY x.created_at DESC, x.id DESC LIMIT 1) f ON true
WHERE a.account_id = @account_id
  AND (NOT @has_cursor::boolean OR (a.created_at, a.id) < (@cursor_created_at::timestamptz, @cursor_id::uuid))
ORDER BY a.created_at DESC, a.id DESC
LIMIT @page_size;

-- name: GetApplicationOfAccount :one
SELECT a.id, a.kind, a.edrpou, a.association_id, a.state, a.created_at,
       f.association_name, f.created_at AS filed_at,
       (SELECT d.reason FROM application_decisions d
         WHERE d.application_id = a.id AND d.outcome = 'rejected'
         ORDER BY d.created_at DESC, d.id DESC LIMIT 1) AS rejection_reason
FROM applications a
JOIN LATERAL (SELECT x.association_name, x.created_at FROM application_filings x
               WHERE x.application_id = a.id ORDER BY x.created_at DESC, x.id DESC LIMIT 1) f ON true
WHERE a.id = @id AND a.account_id = @account_id;

-- name: LockApplicationOfAccount :one
SELECT id, kind, edrpou, state, created_at FROM applications
WHERE id = @id AND account_id = @account_id
FOR UPDATE;

-- name: ListFilings :many
SELECT f.id, f.association_name, f.association_address, f.applicant_full_name, f.position,
       f.term_start, f.term_end, f.created_at,
       e.id AS extract_id, e.media_type, e.original_name, e.sha256, e.size_bytes
FROM application_filings f
JOIN application_extracts e ON e.id = f.extract_id
WHERE f.application_id = @application_id
ORDER BY f.created_at DESC, f.id DESC;

-- name: ListDecisions :many
SELECT id, filing_id, association_id, outcome, reason, relied_on_document, created_at
FROM application_decisions
WHERE application_id = @application_id
ORDER BY created_at DESC, id DESC;

-- name: GetExtract :one
SELECT e.original_name, e.media_type, e.content, e.sha256
FROM application_filings f
JOIN applications a ON a.id = f.application_id
JOIN application_extracts e ON e.id = f.extract_id
WHERE f.id = @filing_id AND a.id = @application_id AND a.account_id = @account_id;

-- name: UpdateApplicationState :exec
UPDATE applications SET state = @state WHERE id = @id;
