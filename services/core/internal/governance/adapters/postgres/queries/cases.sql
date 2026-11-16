-- name: InsertCase :one
INSERT INTO cases (id, association_id, title, created_by_member_id)
VALUES (sqlc.arg(id), sqlc.arg(association_id), sqlc.arg(title), sqlc.arg(created_by_member_id))
RETURNING id, title, created_by_member_id, created_at;

-- name: ListCases :many
-- Pages newest first by the (created_at, id) keyset; a NULL after_created_at starts from the top.
SELECT id, title, created_by_member_id, created_at FROM cases
WHERE association_id = sqlc.arg(association_id)
  AND (sqlc.narg(after_created_at)::timestamptz IS NULL
   OR (created_at, id) < (sqlc.narg(after_created_at)::timestamptz, sqlc.narg(after_id)::uuid))
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(row_limit);

-- name: GetCase :one
SELECT id, title, created_by_member_id, created_at FROM cases
WHERE association_id = sqlc.arg(association_id) AND id = sqlc.arg(id);

-- name: ListCaseDocuments :many
SELECT d.id, d.case_id, d.previous_version_id, d.document_date, d.created_by_member_id, d.created_at,
       f.id AS file_id, f.original_name, f.media_type, f.size_bytes, f.sha256,
       f.uploaded_by_member_id, f.created_at AS file_created_at
FROM case_documents d
JOIN files f ON f.id = d.file_id AND f.association_id = d.association_id
WHERE d.association_id = sqlc.arg(association_id) AND d.case_id = sqlc.arg(case_id)
ORDER BY d.created_at, d.id;

-- name: GetCaseDocument :one
SELECT d.id, d.case_id, d.previous_version_id, d.document_date, d.created_by_member_id, d.created_at,
       f.id AS file_id, f.original_name, f.media_type, f.size_bytes, f.sha256,
       f.uploaded_by_member_id, f.created_at AS file_created_at
FROM case_documents d
JOIN files f ON f.id = d.file_id AND f.association_id = d.association_id
WHERE d.association_id = sqlc.arg(association_id) AND d.id = sqlc.arg(id);

-- name: CaseHasDocument :one
SELECT EXISTS (
    SELECT 1 FROM case_documents
    WHERE association_id = sqlc.arg(association_id) AND case_id = sqlc.arg(case_id) AND id = sqlc.arg(id)
)::bool;

-- name: InsertFile :exec
INSERT INTO files (id, association_id, drive_file_id, drive_revision_id, sha256, size_bytes, media_type, original_name, uploaded_by_member_id)
VALUES (sqlc.arg(id), sqlc.arg(association_id), sqlc.arg(drive_file_id), sqlc.arg(drive_revision_id), sqlc.arg(sha256),
        sqlc.arg(size_bytes), sqlc.arg(media_type), sqlc.arg(original_name), sqlc.arg(uploaded_by_member_id));

-- name: InsertCaseDocument :exec
INSERT INTO case_documents (id, association_id, case_id, file_id, previous_version_id, document_date, created_by_member_id)
VALUES (sqlc.arg(id), sqlc.arg(association_id), sqlc.arg(case_id), sqlc.arg(file_id), sqlc.narg(previous_version_id),
        sqlc.narg(document_date), sqlc.arg(created_by_member_id));
