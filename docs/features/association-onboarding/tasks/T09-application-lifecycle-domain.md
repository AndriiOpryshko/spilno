---
id: T09
title: "Model the application lifecycle and the guards of a decision"
layer: "domain"
deps: ["T03"]
acs: ["AC-03", "AC-03b", "AC-06", "AC-12", "AC-13", "AC-15", "AC-16c", "AC-18b"]
files_hint:
  - "services/core/internal/identity/domain/application.go"
  - "services/core/internal/identity/domain/application_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T09 — Model the application lifecycle and the guards of a decision

## Why

[PRD](../PRD.md) §10 «States», «Allowed transitions», «Refused transitions»; [ADR-0009](../adr/0009-treat-the-operator-role-and-a-roster-membership-as-independent-grants-on-one-account.md) for the own-filing guard; [sad](../sad.md) §8 «Lifecycle».

**Satisfies:** [PRD §5](../PRD.md) AC-03, AC-03b, AC-06, AC-12, AC-13, AC-15, AC-16c, AC-18b · **Blocked by:** [T03](./T03-identity-problem-codes.md)

## What

- States `pending`, `approved`, `rejected`; kinds `connection`, `transfer`.
- `Approve`: refused for the acting оператор's own filing (`identity.operator_own_application`), for a decided application (`identity.application_not_pending`), without a recorded check or with a signature found invalid (`identity.signature_not_checked`), when the extract's ЄДРПОУ differs from the typed or the association's one (`identity.extract_code_mismatch`), and over an open розбіжність імен without a document reference of up to 500 characters (`identity.mismatch_needs_document`).
- `Reject`: the same first two guards, then an empty reason is `identity.reason_required`.
- `Refile`: only from `rejected` and only by the same applicant account; the earlier check is not carried over.

## Definition of Done

- [ ] Acceptance: Unit tests cover every allowed and refused transition of PRD §10 for a заявка and a запит, each refusal with its catalogue code.
- [ ] table test with one case per row of PRD §10 «Allowed transitions» and «Refused transitions» that concerns a заявка or a запит
- [ ] the guards answer in the order listed above when several fail at once
- [ ] the own-filing guard compares accounts and refuses both approve and reject
- [ ] `mise run lint` and `mise run test` green

## Notes

The comparison of the own filing is by account, not by person (SAD §11 accepted risk). Recording a check is not guarded: it leads to no decision (ADR-0009).
