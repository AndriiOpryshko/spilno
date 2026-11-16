---
id: T56
title: "Send a valid Content-Disposition on the review download and tighten the extract type check"
layer: "ports"
deps: []
acs: ["AC-02", "AC-03"]
files_hint:
  - "services/core/internal/identity/adapters/http/review.go"
  - "services/core/internal/identity/domain/extract_file.go"
  - "services/core/internal/identity/domain/extract_file_test.go"
  - "services/core/internal/identity/review_test.go"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T56 — Send a valid Content-Disposition on the review download and tighten the extract type check

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), B5, B6. [ADR-0005](../adr/0005-check-the-signed-extract-manually-and-keep-the-uploaded-file-in-identity.md): the оператор downloads the file to check it; [PRD](../PRD.md) §6: the file type is content-sniffed. The review download quotes the name with `%q`, and the sniff accepts any file whose first byte is ASCII `0`.

**Satisfies:** [PRD §5](../PRD.md) AC-02, AC-03 · **Blocked by:** —

## What

- `mime.FormatMediaType("attachment", …)` in `DownloadExtractForReview`, as in `applications.go` and `card.go`.
- A DER container needs the SEQUENCE tag and a valid length byte (and the signedData OID prefix for PKCS#7).

## Definition of Done

- [ ] Acceptance: The review download of a file with a Cyrillic name carries the same RFC 6266 header as the two other downloads, and a text file that starts with `0` is refused as an extract.
- [ ] handler test with the name `витяг.pdf.p7s`
- [ ] domain tests: a real `.p7s` prefix accepted, `0123…` text and `0<html>` refused
- [ ] `mise run lint` and `mise run test` green
