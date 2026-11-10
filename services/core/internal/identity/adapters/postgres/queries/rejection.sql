-- name: LockApplicationForRejection :one
SELECT account_id, kind, edrpou, state FROM applications WHERE id = @id FOR UPDATE;

-- name: CreateRejectionDecision :one
INSERT INTO application_decisions (id, application_id, filing_id, outcome, reason, decided_by_email)
VALUES (@id, @application_id, @filing_id, 'rejected', @reason, @decided_by_email)
RETURNING created_at;
