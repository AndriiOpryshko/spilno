# Staged migrations — association-onboarding

Staged, not live. Promote in this order with `mise run migrate:new <name>` (never hand-type the number),
one PR per «Ships in» group, each after the previous one is deployed
([goose practice](../../../../.claude/skills/ao-data-model/references/goose-practice.md#promote)).
Checked 2026-11-05 on a throwaway `postgres:18.6-alpine` after live `00027`: up → down-to 27 → up,
schema diff empty except the column position of `accounts.association_id` after Down; sqlc compile OK.

| Order | Staged file | Ships in | Depends on | Expand/contract phase |
|---|---|---|---|---|
| 1 | `NNNNN_create_applications.sql` | PR 1 | - | - |
| 2 | `NNNNN_add_associations_onboarding_columns.sql` | PR 1 | 1 | expand (`state`) |
| 3 | `NNNNN_index_associations_edrpou.sql` | PR 1 | 2 | - (NO TRANSACTION) |
| 4 | `NNNNN_index_associations_confirmed_by_decision.sql` | PR 1 | 2 | - (NO TRANSACTION) |
| 5 | `NNNNN_create_association_state_changes.sql` | PR 1 | - | - |
| 6 | `NNNNN_add_board_member_revisions_position.sql` | PR 1 | - | expand (`position`) |
| 7 | `NNNNN_index_board_member_revisions_sign_in_email.sql` | PR 1 | - | - (NO TRANSACTION) |
| 8 | `NNNNN_add_contracts_association_id.sql` | PR 1 | - | expand |
| 9 | `NNNNN_create_app_role.sql` | PR 1 | 1, 5 | - |
| 10 | `NNNNN_backfill_associations_state.sql` | PR 2 | 2 | backfill — [note](./backfill-associations-state.md) |
| 11 | `NNNNN_backfill_board_member_revisions_position.sql` | PR 2 | 6 | backfill — [note](./backfill-board-member-revisions-position.md) |
| 12 | `NNNNN_backfill_contracts_association_id.sql` | PR 2 | 8 | backfill — [note](./backfill-contracts-association-id.md) |
| 13 | `NNNNN_contract_associations_state.sql` | PR 3 | 10 | contract |
| 14 | `NNNNN_contract_board_member_revisions_position.sql` | PR 3 | 11 | contract |
| 15 | `NNNNN_contract_contracts_association_id.sql` | PR 3 | 12 | contract |
| 16 | `NNNNN_index_contracts_association_created.sql` | PR 3 | 15 | - (NO TRANSACTION) |
| 17 | `NNNNN_enable_rls_governance.sql` | PR 3 | 9 | - |
| 18 | `NNNNN_enable_rls_finance.sql` | PR 3 | 9, 15 | - |
| 19 | `NNNNN_enable_rls_identity.sql` | PR 3 | 9 | - |
| 20 | `NNNNN_drop_accounts_association_id.sql` | PR 4 | the release that stops reading it | contract (drop column) |

PR 3 ships together with the app change that connects as a login role in `spilno_app` and runs every
association-scoped request through `platform/db.InAssociation` (ADR-0003); before that change, enabling
row-level security would hide every row from the app.
