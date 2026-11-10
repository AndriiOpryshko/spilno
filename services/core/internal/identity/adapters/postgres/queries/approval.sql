-- name: GetApplicationForApproval :one
SELECT a.id, a.kind, a.account_id, a.edrpou, a.state, a.association_id, c.email AS applicant_email
FROM applications a
JOIN accounts c ON c.id = a.account_id
WHERE a.id = @id
FOR UPDATE OF a;

-- name: GetLatestFilingForApproval :one
SELECT id, association_name, association_address, applicant_full_name, term_start, created_at
FROM application_filings
WHERE application_id = @application_id
ORDER BY created_at DESC, id DESC
LIMIT 1;

-- name: CreateAssociation :exec
INSERT INTO associations (id, name, edrpou, address, state)
VALUES (@id, @name, @edrpou, @address, 'connected');

-- name: CreateApprovalDecision :one
INSERT INTO application_decisions (
    id, application_id, filing_id, check_id, association_id, outcome, relied_on_document, decided_by_email
) VALUES (
    @id, @application_id, @filing_id, @check_id, @association_id, 'approved', @relied_on_document, @decided_by_email
)
RETURNING created_at;

-- name: SetApplicationApproved :exec
UPDATE applications SET state = 'approved', association_id = @association_id WHERE id = @id;

-- name: SetAssociationConfirmation :exec
UPDATE associations SET confirmed_by_decision_id = @decision_id WHERE id = @id;

-- name: GetApplicationScope :one
SELECT kind, association_id FROM applications WHERE id = @id;

-- name: GetAssociationForApproval :one
SELECT edrpou, confirmed_by_decision_id FROM associations WHERE id = @id FOR UPDATE;

-- name: GetEntryOfDecision :one
SELECT board_member_id, COALESCE(sign_in_email, '')::text AS sign_in_email, term_start, term_end
FROM (
    SELECT DISTINCT ON (board_member_id) board_member_id, sign_in_email, term_start, term_end, decision_id
    FROM board_member_revisions
    WHERE association_id = @association_id
    ORDER BY board_member_id, created_at DESC, id DESC
) latest
WHERE decision_id = @decision_id AND term_end IS NULL
LIMIT 1;

-- name: SetAssociationEdrpou :exec
UPDATE associations SET edrpou = @edrpou WHERE id = @id AND edrpou IS NULL;
