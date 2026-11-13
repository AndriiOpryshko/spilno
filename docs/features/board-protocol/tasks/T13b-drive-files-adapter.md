---
id: T13b
title: "Implement the Files port on Google Drive and pass the shared contract on a fake Drive"
layer: "infra"
parent: T13
acs: ["AC-26", "AC-14", "AC-17"]
scope: governance
package: services/core/internal/governance/adapters/gdrive
test_cmd: "cd services/core && go test -count=1 ./internal/governance/adapters/gdrive/"
stub: services/core/internal/governance/adapters/gdrive/files.go
test_files:
  - services/core/internal/governance/adapters/gdrive/files_test.go
  - services/core/internal/governance/adapters/gdrive/drivefake_test.go
estimate: "a day"
status: "in_progress"
---

# T13b — Implement the Files port on Google Drive

Slice of [T13](./T13-drive-files-adapter.md); [ADR-0005](../adr/0005-store-files-in-association-owned-google-drive-folder.md):
an association's files live in its own Drive folder; spilno computes the sha256 while streaming and compares it
with Drive's `sha256Checksum`, keeps the head revision id, sets `contentRestrictions.readOnly` on pinned files
and restores a trashed one with `files.untrash`. [T13a](./T13a-files-port-fake.md) declared the port
(`governance/ports/files.go`) and the contract (`governance/adapters/filestest`); this slice is the adapter that
passes the same contract against a fake Drive served by `httptest`.

Not here: building the HTTP client of an association from `identity.GoogleTokenSource`, finding its folder id and
handing the adapter to the governance use cases (with the first use case that stores a file, T21); a reason per
restriction («protocol locked», «protocol withdrawn»), which needs a port change; resuming an interrupted upload;
running the contract against real Google, which needs a test account and a manual consent.

## Interface

Package `gdrive` (`services/core/internal/governance/adapters/gdrive`), file `files.go`:

```go
// ReadOnlyReason is the reason Drive shows on a file spilno made read-only.
const ReadOnlyReason = "locked by spilno"

// Config is one association's Drive storage.
type Config struct {
    HTTP      *http.Client // authorized as the association's Google account
    APIURL    string       // "https://www.googleapis.com/drive/v3" in production; no trailing slash
    UploadURL string       // "https://www.googleapis.com/upload/drive/v3" in production; no trailing slash
    FolderID  string       // the association's spilno folder; every upload goes there
}

// Files is ports.Files on Google Drive.
type Files struct{ /* unexported */ }

var _ ports.Files = (*Files)(nil)

func New(cfg Config) *Files

func (f *Files) Upload(ctx context.Context, name, mediaType string, body io.Reader) (ports.StoredFile, error)
func (f *Files) Open(ctx context.Context, id string) (io.ReadCloser, error)
func (f *Files) Stat(ctx context.Context, id string) (ports.FileStatus, error)
func (f *Files) SetReadOnly(ctx context.Context, id string) error
func (f *Files) Untrash(ctx context.Context, id string) error
```

No new module: plain `net/http` and `encoding/json`. The package imports `governance/ports` and
`governance/domain`, nothing from another context.

## Rules

- Every request goes through `cfg.HTTP` with the caller's context.
- Drive answers only `kind`, `id`, `name` and `mimeType` unless `fields` asks for more, so every request whose
  answer the adapter reads sends `fields` with what it needs out of `id`, `headRevisionId`, `sha256Checksum`,
  `size`, `mimeType`, `trashed`, `contentRestrictions`. `size` arrives as a decimal string.
- `Upload` is a resumable upload in one piece:
  1. `POST {UploadURL}/files?uploadType=resumable` with the JSON metadata
     `{"name": name, "mimeType": mediaType, "parents": [FolderID]}`; the answer's `Location` header is the session
     URL.
  2. `PUT <session URL>` with `Content-Type: mediaType` and the body streamed, never buffered whole; the sha256 and
     the size are computed while it streams. The answer is the file.
  3. Drive's `sha256Checksum` must equal the computed one. Otherwise the adapter deletes the file
     (`DELETE {APIURL}/files/{id}`) and returns `ports.ErrChecksumMismatch`.
  The result is `ID` = `id`, `Revision` = `headRevisionId`, `SHA256` = the computed lowercase hex, `Size` = the
  bytes streamed, `MediaType` = `mediaType`.
- `Open` is `GET {APIURL}/files/{id}?alt=media`; the answer's body is returned as it is and the caller closes it.
- `Stat` is `GET {APIURL}/files/{id}`. `ReadOnly` is true when any item of `contentRestrictions` has
  `readOnly: true`.
- `SetReadOnly` is `PATCH {APIURL}/files/{id}` with `{"contentRestrictions":[{"readOnly":true,"reason":ReadOnlyReason}]}`.
- `Untrash` is `PATCH {APIURL}/files/{id}` with `{"trashed":false}`.
- Errors, for every method, on a failure nothing is returned but the error (zero value, nil reader):
  - the caller's context is cancelled or past its deadline → the context's error (`errors.Is`);
  - the upload body fails to read → an error that wraps the body's error;
  - Drive answers `404` → `ports.ErrFileNotFound`;
  - Drive answers `401`, `403`, `408`, `429` or `5xx`, or the request cannot be sent or its answer cannot be read
    for another reason → `*domain.DomainError{Code: domain.CodeStorageUnavailable}`;
  - any other answer outside `2xx`, an answer that is not the expected JSON, or a missing `Location`, `id`,
    `headRevisionId` or `sha256Checksum` → a plain error whose text has the HTTP method, the path and the status,
    and is neither of the two above.
- Tests are in package `gdrive`. `drivefake_test.go` holds `driveFake`, an `httptest.Server` that behaves like
  Drive for these calls:
  - it serves `/drive/v3` and `/upload/drive/v3`, and answers `401` to a request without
    `Authorization: Bearer at-1`; the tests' HTTP client adds that header in a `RoundTripper`;
  - the session `POST` answers `200` with `Location: <server>/upload/drive/v3/files?uploadType=resumable&upload_id=<n>`;
    the `PUT` stores the file as `id` `file-<n>`, `headRevisionId` `rev-<n>` with the sha256 and the size of the
    bytes it received and the metadata of the session;
  - `GET`, `PATCH` and `DELETE` on `/drive/v3/files/{id}` read (`alt=media` gives the bytes), change `trashed`
    and `contentRestrictions`, and remove; an unknown id is `404`;
  - like Drive, an answer holds only `kind`, `id`, `name` and `mimeType` plus the top-level names `fields` lists
    (for the `PUT`, the `fields` of its session `POST`);
  - it records every request (method, path, query, `Content-Type`, body) and offers the test `trash(id)`,
    `len()`, `file(id)`, `reportWrongSHA256(on)`, `answerEverything(status)` and `close()`.

## Acceptance criteria

- **AC-1** Given a new fake Drive per subtest and an adapter on it with the folder `folder-1`, Then
  `filestest.Run` passes, with the fake's `trash` as the harness's `Trash`.
- **AC-2** Given that adapter, When `scan.pdf` of 1000 bytes is uploaded as `application/pdf`, Then Drive received
  exactly two requests: `POST /upload/drive/v3/files` with `uploadType=resumable` and a JSON body whose `name` is
  `scan.pdf`, `mimeType` is `application/pdf` and `parents` is `["folder-1"]`; then `PUT /upload/drive/v3/files`
  with the `upload_id` of the session, `Content-Type: application/pdf` and exactly those bytes. The result's `ID`
  and `Revision` are the `id` and `headRevisionId` of the file the fake holds.
- **AC-3** Given the fake reports a wrong sha256, When a file is uploaded, Then the error is
  `ports.ErrChecksumMismatch`, the result is the zero `StoredFile`, the fake holds no file and its last request
  was `DELETE /drive/v3/files/<the id it had given>`; and with the fault off the same upload succeeds and the fake
  holds one file.
- **AC-4** Given a stored file that the storage trashed, When Drive answers `503` to everything, and likewise
  `401`, `403`, `429` and `500`, and likewise when its server is closed, Then `Upload`, `Open`, `Stat`,
  `SetReadOnly` and `Untrash` each return a `*domain.DomainError` with the code
  `governance.storage_unavailable`; and once Drive answers again it still holds one file, trashed, not read-only,
  with the same sha256.
- **AC-5** Given a stored file, Then `SetReadOnly` sends `PATCH /drive/v3/files/<id>` whose JSON body has one
  `contentRestrictions` item with `readOnly` true and `reason` `locked by spilno`; `Untrash` sends
  `PATCH /drive/v3/files/<id>` whose JSON body has `trashed` false; `Stat` sends `GET /drive/v3/files/<id>` whose
  `fields` names `headRevisionId`, `sha256Checksum`, `size`, `mimeType`, `trashed` and `contentRestrictions`; and
  `Open` sends `GET /drive/v3/files/<id>` with `alt=media`.
- **AC-6** Given a stored file and Drive answering `400` to everything, Then each of the five methods returns an
  error that is neither `ports.ErrFileNotFound` nor a `*domain.DomainError`, and whose text contains `400`; and
  afterwards the fake holds one file.
- **AC-7** Given a stored file and a cancelled context, Then `Open`, `Stat`, `SetReadOnly` and `Untrash` each
  return an error that matches `context.Canceled` and is not a `*domain.DomainError`.
