---
id: T08
title: "Validate a filing: every field at once, the term rules and the extract file"
layer: "domain"
deps: ["T03"]
acs: ["AC-01", "AC-02", "AC-30"]
files_hint:
  - "services/core/internal/identity/domain/filing.go"
  - "services/core/internal/identity/domain/filing_test.go"
  - "services/core/internal/identity/domain/extract_file.go"
  - "services/core/internal/identity/domain/extract_file_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T08 — Validate a filing: every field at once, the term rules and the extract file

## Why

[PRD](../PRD.md) AC-02, AC-30 and §6 «File size»; [sad](../sad.md) §8 «Validation», «Files», «Dates»; [test-plan](../test-plan.md) rows AC-02, AC-30 and «File size and kind».

**Satisfies:** [PRD §5](../PRD.md) AC-01, AC-02, AC-30 · **Blocked by:** [T03](./T03-identity-problem-codes.md)

## What

- `Filing` (kind, ЄДРПОУ, association name and address, the applicant's full name, position, term start and end) with `Validate(filingDate)`: ЄДРПОУ of 8 digits, name, address and full name present and within the bounds of [data-model](../data-model.md), position `head` only, term start present and not later than the filing date, an extract attached — all broken fields in one `identity.application_invalid`.
- A term end earlier than the filing date is `identity.term_ended`; a term that ends on the filing date and a missing term end are accepted.
- `ExtractFile`: one file ≤ 5 MB whose first bytes are a PDF, a `.p7s` container or an ASiC-E container; anything else is a broken field.
- Dates are calendar dates in Europe/Kyiv from an injected clock.

## Definition of Done

- [ ] Acceptance: Table tests name every broken field of a заявка and of a запит in one answer, refuse a term that ended before the filing date, and accept an extract only as one PDF or signed container of at most 5 MB.
- [ ] table test with the cases of the test plan row AC-02, for both kinds
- [ ] table test: end before the filing date refused; end on the filing date accepted; no end accepted
- [ ] exactly 5 MB accepted, one byte more refused, a file that is neither a PDF nor a signed container refused
- [ ] `mise run lint` and `mise run test` green

## Notes

Stdlib only (`domain/`). The signature is never checked here (PRD §10 States).
