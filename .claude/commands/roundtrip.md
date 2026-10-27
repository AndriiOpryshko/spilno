---
description: Schema roundtrip on the local DB - baseline N, up, down-to N, up - and diff pg_dump --schema-only
argument-hint: <N, the last migration version that must survive, e.g. 7>
allowed-tools: Bash(.claude/scripts/roundtrip.sh *), Bash(ls services/core/migrations*)
---

Prove that the migrations after version `$ARGUMENTS` are reversible: every Down restores the
version-N schema exactly and a second Up rebuilds the same latest schema.

1. No argument: list `services/core/migrations/` and ask which version to use as N (usually the
   last migration on `main`, so only this branch's migrations roll back). Stop.
2. Warn in one line that `down-to N` deletes local data above version N, then run
   `.claude/scripts/roundtrip.sh $ARGUMENTS` from the repo root (timeout 5 minutes). It starts
   the infra, refuses while the core API holds :8080 (it would re-apply migrations on start),
   drops the `\restrict` lines pg_dump 18 prints and exits non-zero on any schema diff.
3. Report: the versions walked, `OK` / `FAIL` for down and up, and for a FAIL the diff hunks
   reduced to the objects that differ (table, column, index, constraint, trigger) and which
   migration's Down (or Up) most likely misses them. Do not edit migrations unless asked.
4. On `feature-1/*` / `feature-2/*` branches append to `LOG.md`
   `- <date +%Y-%m-%dT%H:%M:%S%z> | <F1|F2> | check | schema roundtrip v$ARGUMENTS -> <pass|FAIL: what>`.
