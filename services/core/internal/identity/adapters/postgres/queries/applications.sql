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
