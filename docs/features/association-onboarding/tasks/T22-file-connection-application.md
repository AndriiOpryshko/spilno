---
id: T22
title: "File a заявка на підключення and keep the signed extract unchanged"
layer: "ports"
deps: ["T05", "T08", "T12", "T18"]
acs: ["AC-01", "AC-02", "AC-04", "AC-04b", "AC-04c", "AC-30"]
files_hint:
  - "services/core/internal/identity/app/"
  - "services/core/internal/identity/ports/ports.go"
  - "services/core/internal/identity/adapters/postgres/"
  - "services/core/internal/identity/adapters/http/"
  - "services/core/internal/identity/identitytest/"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T22 — File a заявка на підключення and keep the signed extract unchanged

## Why

[sad](../sad.md) §6 Flow 1; [ADR-0005](../adr/0005-check-the-signed-extract-manually-and-keep-the-uploaded-file-in-identity.md): the uploaded file is kept in `identity`, unchanged, with its sha256; [data-model](../data-model.md) tables `applications`, `application_filings`, `application_extracts` and the queries `CreateExtract`, `CreateApplication`, `CreateFiling`, `GetAssociationByEdrpou`, `FindOpenApplicationOfAccount`.

**Satisfies:** [PRD §5](../PRD.md) AC-01, AC-02, AC-04, AC-04b, AC-04c, AC-30 · **Blocked by:** [T05](./T05-promote-expand-migrations.md), [T08](./T08-filing-validation-domain.md), [T12](./T12-two-database-roles.md), [T18](./T18-sign-in-any-verified-person.md)

## What

- sqlc queries and repositories for the three tables; the use case in one transaction: validate (T08) → a connected code is `identity.association_already_connected` and names no person → an open заявка of the same account for the code is `identity.application_already_open` → insert the extract (content, sha256, size, sniffed media type, original name), the application `pending` and its first filing.
- Handler `FileApplication` for `multipart/form-data` with `kind = connection`; the body is capped at the file limit plus the form.
- Another person's pending заявка for the same code does not block and is not revealed (AC-04c).

## Definition of Done

- [ ] Acceptance: fileApplication records a complete заявка as pending with its filing and the extract byte for byte, and answers every refusal of AC-02, AC-04, AC-04b and AC-30 with its code while storing nothing.
- [ ] integration + contract: a complete filing → 201 `pending` with the filing date; the stored bytes and sha256 equal the upload; the same for a person who is a член правління of another association
- [ ] each case of AC-02 → 422 with every broken field in `errors[]`; the row counts of the three tables are unchanged
- [ ] AC-04 → 409 without any person in the body; AC-04b → 409; a second person for the same code → 201 of the same shape as a lone applicant; AC-30 → 409
- [ ] a 5 MB fixture is filed in under 5 s; one byte more is refused and nothing is stored
- [ ] `mise run lint` and `mise run test` green

## Notes

Head of the identity lane (T22 → T33 share `adapters/postgres` and its generated `sqlcgen`). Extract fixtures are a few bytes with a PDF header; no real name or code (`.claude/rules/openapi.md`).
