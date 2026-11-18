-- name: GetApplicationForApproval :one
SELECT a.id, a.kind, a.account_id, a.edrpou, a.state, a.association_id, c.email AS applicant_email
FROM applications a
JOIN accounts c ON c.id = a.account_id
WHERE a.id = @id
FOR UPDATE OF a;

-- name: GetLatestFilingForApproval :one
SELECT id, association_name, association_address, applicant_full_name, term_start, term_end, created_at
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
-- The entry a decision seated is the board member whose first revision carrying the decision is the latest;
-- a later edit copies the decision into more revisions but never moves that first one.
WITH firsts AS (
    SELECT DISTINCT ON (f.board_member_id) f.board_member_id, f.created_at, f.id
    FROM board_member_revisions f
    WHERE f.association_id = @association_id AND f.decision_id = @decision_id
    ORDER BY f.board_member_id, f.created_at, f.id
), seated AS (
    SELECT board_member_id FROM firsts ORDER BY created_at DESC, id DESC LIMIT 1
)
SELECT l.board_member_id, COALESCE(l.sign_in_email, '')::text AS sign_in_email, l.term_start, l.term_end
FROM seated s
JOIN LATERAL (
    SELECT r.board_member_id, r.sign_in_email, r.term_start, r.term_end
    FROM board_member_revisions r
    WHERE r.association_id = @association_id AND r.board_member_id = s.board_member_id
    ORDER BY r.created_at DESC, r.id DESC
    LIMIT 1
) l ON true;

-- name: SetAssociationEdrpou :exec
UPDATE associations SET edrpou = @edrpou WHERE id = @id AND edrpou IS NULL;
