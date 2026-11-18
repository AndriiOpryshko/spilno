# Staged migrations — association-onboarding

Promoted: files 1–21 are live as `00028`–`00048` (PR #48, 2026-11-10) and file 22 as `00049`; the files here
are the frozen design record. They were written to be promoted in this order with `mise run migrate:new <name>`
(never hand-type the number), one PR per planned group, each after the previous one is deployed
([goose practice](../../../../.claude/skills/ao-data-model/references/goose-practice.md#promote)); all four groups
shipped in one PR instead. What that means and why it was safe:
[data-model, «Delivery: one release instead of four»](../data-model.md#delivery-one-release-instead-of-four).
Checked 2026-11-05 on a throwaway `postgres:18.6-alpine` after live `00027`: up → down-to 27 → up,
schema diff empty except the column position of `accounts.association_id` after Down; sqlc compile OK.

| Order | Staged file | Ships in | Depends on | Expand/contract phase |
|---|---|---|---|---|
| 1 | `NNNNN_create_applications.sql` | #48 (planned: PR 1) | - | - |
| 2 | `NNNNN_add_associations_onboarding_columns.sql` | #48 (planned: PR 1) | 1 | expand (`state`) |
| 3 | `NNNNN_index_associations_edrpou.sql` | #48 (planned: PR 1) | 2 | - (NO TRANSACTION) |
| 4 | `NNNNN_index_associations_confirmed_by_decision.sql` | #48 (planned: PR 1) | 2 | - (NO TRANSACTION) |
| 5 | `NNNNN_create_association_state_changes.sql` | #48 (planned: PR 1) | - | - |
| 6 | `NNNNN_add_board_member_revisions_position.sql` | #48 (planned: PR 1) | - | expand (`position`) |
| 7 | `NNNNN_index_board_member_revisions_sign_in_email.sql` | #48 (planned: PR 1) | - | - (NO TRANSACTION) |
| 8 | `NNNNN_add_contracts_association_id.sql` | #48 (planned: PR 1) | - | expand |
| 9 | `NNNNN_create_app_role.sql` | #48 (planned: PR 1) | 1, 5 | - |
| 10 | `NNNNN_drop_accounts_association_id.sql` | #48 (planned: PR 4) | the release that stops reading it (already out: no Go code touches `accounts.association_id`) | contract (drop column) - promoted early, see below |
| 11 | `NNNNN_backfill_associations_state.sql` | #48 (planned: PR 2) | 2 | backfill — [note](./backfill-associations-state.md) |
| 12 | `NNNNN_backfill_board_member_revisions_position.sql` | #48 (planned: PR 2) | 6 | backfill — [note](./backfill-board-member-revisions-position.md) |
| 13 | `NNNNN_backfill_contracts_association_id.sql` | #48 (planned: PR 2) | 8 | backfill — [note](./backfill-contracts-association-id.md) |
| 14 | `NNNNN_contract_associations_state.sql` | #48 (planned: PR 3) | 11 | contract |
| 15 | `NNNNN_contract_board_member_revisions_position.sql` | #48 (planned: PR 3) | 12 | contract |
| 16 | `NNNNN_contract_contracts_association_id.sql` | #48 (planned: PR 3) | 13 | contract |
| 17 | `NNNNN_index_contracts_association_created.sql` | #48 (planned: PR 3) | 16 | - (NO TRANSACTION) |
| 18 | `NNNNN_enable_rls_governance.sql` | #48 (planned: PR 3) | 9 | - |
| 19 | `NNNNN_enable_rls_finance.sql` | #48 (planned: PR 3) | 9, 16 | - |
| 20 | `NNNNN_enable_rls_identity.sql` | #48 (planned: PR 3) | 9 | - |
| 21 | `NNNNN_enable_rls_identity_state_changes.sql` | #48 (planned: PR 3) | 5, 9 | - |
| 22 | `NNNNN_index_board_member_revisions_lower_sign_in_email.sql` | review follow-up (T59) | - | - (NO TRANSACTION) |
| 23 | `NNNNN_index_files_association_id.sql` | second re-review follow-up (T114) | - | - (NO TRANSACTION) |

The planned PR 3 was to ship together with the app change that connects as a login role in `spilno_app` and runs every
association-scoped request through `platform/db.InAssociation` (ADR-0003); before that change, enabling
row-level security would hide every row from the app.

**Order changed (T06):** row 10 was row 20. `accounts.association_id` is `NOT NULL`, so a person off every roster
cannot get an account row (AC-05) until it is dropped; and the drop was last only «after the release that stops
reading it», which is already out (no Go code reads or writes it). Left last, the order loops: RLS (T20) needs the
middleware (T19), which needs sign-in (T18), which needs this drop. Dependencies of the later rows were renumbered.

File 21 was added during implementation (T20, 2026-11-10): the isolation test walks every table with an
`association_id` column and found `association_state_changes` without a policy. `applications` and
`application_decisions` stay without one on purpose (ADR-0003: no association before approval, reached only
by account or by the оператор) and are the two named exceptions of that test.

File 22 was added by the review follow-up (T59, 2026-11-16; live `00049`): `MembershipsOf` and the own-entry
clause of the row-level security policy compare `lower(sign_in_email)`, which the plain index of file 7 cannot
serve, so the lookup was a sequential scan. The index of file 7 stays until a later release drops it.

File 23 was added by the second re-review follow-up (T114, 2026-11-18; live `00050`): refusing another storage
account while the association has files asks `files` by `association_id`, and no index of `files` led with it.
