export const meta = {
  name: 'openapi-contract-audit',
  description: 'Read-only audit of every operation in services/core/api/openapi.yaml against spilno openapi rules, with independent verification',
  whenToUse: 'Re-audit the live OpenAPI contract against .claude/rules/openapi.md after contract or handler changes. args: {cap} overrides the agent ceiling (default 140 = 1 list + 41 audits + ~92 verifies + 1 report).',
  phases: [
    { title: 'List', detail: 'one agent collects all operationIds' },
    { title: 'Audit', detail: 'sonnet auditor per operation, up to 2 retries' },
    { title: 'Verify', detail: 'independent verifier per finding' },
    { title: 'Report', detail: 'write docs/features/board-protocol/workflow/audit.md' },
  ],
}

// First run (wf_27380f86-0c6) used CAP = 90 and left 45 of 92 findings unverified; default raised to fit every finding.
const CAP = (args && args.cap) || 140
let used = 0
const spawn = (p, o) => { used++; return agent(p, o) }

const READONLY = 'STRICTLY READ-ONLY: do not create, edit, or delete any file; do not run git commands that change state; do not run code generation. Work in /Users/andriio/course/learning/spilno-wf-audit.'

const RULES = [
  'R1-tag', 'R2-security', 'R3-auth-errors', 'R4-422-problem-json',
  'R5-code-registry', 'R6-pagination-revision', 'R7-handler',
]
const RULES_TEXT = `
R1-tag: exactly one tag, equal to the path prefix /<context>/.
R2-security: global SessionCookie applies, or explicit security: [] ONLY for public ops (health, sign-in redirect, callback). A public op must not be protected and vice versa.
R3-auth-errors: every protected op declares 401 with code identity.session_required and 403 with code identity.not_current_member.
R4-422-problem-json: an op with a request body declares 422; all error responses use application/problem+json.
R5-code-registry: every error code mentioned in the op's responses exists in the Go registry (services/core/internal/**/domain/errors.go, services/core/internal/platform/httpapi) with the identical string.
R6-pagination-revision: list ops use ?cursor=&limit= and respond {items, next_cursor}; protocol mutations carry protocol_revision and declare 409 governance.draft_revision_conflict. Use status n/a when not applicable.
R7-handler: handler status implemented / stub (returns ErrNotImplemented or similar) / missing in services/core/internal/<bc>/adapters/http; for implemented, check it returns the codes and statuses the contract declares (and none undeclared).`

const LIST_SCHEMA = {
  type: 'object',
  properties: { operations: { type: 'array', items: { type: 'object',
    properties: { operationId: { type: 'string' }, method: { type: 'string' }, path: { type: 'string' }, tags: { type: 'array', items: { type: 'string' } } },
    required: ['operationId', 'method', 'path'] } } },
  required: ['operations'],
}

const AUDIT_SCHEMA = {
  type: 'object',
  properties: {
    operationId: { type: 'string' },
    handler_status: { type: 'string', enum: ['implemented', 'stub', 'missing'] },
    rules: { type: 'array', items: { type: 'object', properties: {
      rule: { type: 'string', enum: RULES }, status: { type: 'string', enum: ['pass', 'fail', 'n/a'] }, note: { type: 'string' } },
      required: ['rule', 'status'] } },
    findings: { type: 'array', items: { type: 'object', properties: {
      rule: { type: 'string', enum: RULES }, claim: { type: 'string' }, evidence: { type: 'string', description: 'file:line, repo-relative' } },
      required: ['rule', 'claim', 'evidence'] } },
  },
  required: ['operationId', 'handler_status', 'rules', 'findings'],
}

const VERDICT_SCHEMA = {
  type: 'object',
  properties: { confirmed: { type: 'boolean' }, evidence: { type: 'string', description: 'file:line you checked' }, reason: { type: 'string' } },
  required: ['confirmed', 'evidence', 'reason'],
}

phase('List')
const listed = await spawn(`${READONLY}
Parse services/core/api/openapi.yaml (e.g. with yq or a careful read) and return EVERY operation: operationId, HTTP method, path, tags. Do not skip any. The user expects about 41.`,
  { phase: 'List', schema: LIST_SCHEMA, effort: 'low', label: 'list-operations' })
const ops = (listed && listed.operations) || []
log(`Operations found: ${ops.length}${ops.length !== 41 ? ' (expected 41!)' : ''}`)

const isValid = r => r && r.operationId && Array.isArray(r.rules) && r.rules.length >= RULES.length - 1 && Array.isArray(r.findings)

const failed = []
let filtered = 0
const unverifiedCap = []

const results = await pipeline(ops,
  async (op) => {
    let attempt = 0
    while (attempt < 3) {
      if (used >= CAP - 1) { log(`Agent cap reached; ${op.operationId} not audited`); break }
      const r = await spawn(`${READONLY}
Audit ONE operation of the live contract services/core/api/openapi.yaml: ${op.method.toUpperCase()} ${op.path} (operationId ${op.operationId}).
Rules come from .claude/rules/openapi.md section "spilno specifics". Check each:
${RULES_TEXT}
Return a status for all 7 rules (pass / fail / n/a with a short note) and the handler status.
For every fail, add a finding {rule, claim, evidence} where evidence is a precise repo-relative file:line (the spec line or Go line that proves it). Only report what you verified by reading the files; no speculation.`,
        { phase: 'Audit', schema: AUDIT_SCHEMA, model: 'sonnet', label: `audit:${op.operationId}${attempt ? ` (retry ${attempt})` : ''}` })
      if (isValid(r)) return { op, audit: r }
      attempt++
      log(`Auditor for ${op.operationId} returned broken/empty result (attempt ${attempt})`)
    }
    failed.push(op.operationId)
    return null
  },
  async (res) => {
    if (!res) return null
    const { op, audit } = res
    const verdicts = await parallel(audit.findings.map(f => async () => {
      if (used >= CAP - 1) { unverifiedCap.push({ operationId: op.operationId, ...f }); return null }
      const v = await spawn(`${READONLY}
Independently check whether this claim about the spilno API contract is TRUE. Be skeptical: if you cannot confirm it from the files, answer confirmed=false.
Operation: ${op.method.toUpperCase()} ${op.path} (operationId ${op.operationId}) in services/core/api/openapi.yaml.
Rule (from .claude/rules/openapi.md, "spilno specifics"): ${f.rule}
Claim: ${f.claim}
Cited location: ${f.evidence}
Read the cited location and anything else needed (spec, services/core/internal/**/domain/errors.go, internal/platform/httpapi, handlers in internal/<bc>/adapters/http). Return the precise file:line that proves or disproves it.`,
        { phase: 'Verify', schema: VERDICT_SCHEMA, label: `verify:${op.operationId}:${f.rule}` })
      return v ? { ...f, verdict: v } : null
    }))
    const confirmed = verdicts.filter(v => v && v.verdict.confirmed)
    filtered += verdicts.filter(v => v && !v.verdict.confirmed).length
    return { operationId: op.operationId, method: op.method, path: op.path, handler_status: audit.handler_status, rules: audit.rules, confirmed }
  },
)

const audited = results.filter(Boolean)
log(`Audited ${audited.length}/${ops.length}, failed ${failed.length}, filtered ${filtered}, unverified-by-cap ${unverifiedCap.length}, agents used ${used}`)

phase('Report')
const payload = JSON.stringify({ total_operations: ops.length, audited, filtered_count: filtered, failed, unverified_due_to_agent_cap: unverifiedCap }, null, 1)
await spawn(`Work in /Users/andriio/course/learning/spilno-wf-audit. Write EXACTLY ONE file: docs/features/board-protocol/workflow/audit.md (create the directory if needed). Do not create or modify any other file, do not commit.
Content (English or Ukrainian prose is fine; keep ids/codes verbatim):
1. Title, scope (live contract services/core/api/openapi.yaml vs .claude/rules/openapi.md "spilno specifics"), method (sonnet auditor per operation, independent verifier per finding, only confirmed kept).
2. Summary counts: operations total, audited, confirmed findings, filtered (unconfirmed) count, failed operations, unverified due to agent cap (if any).
3. Table operation × rule: rows = operationId (with METHOD path), columns = handler status, R1-tag, R2-security, R3-auth-errors, R4-422-problem-json, R5-code-registry, R6-pagination-revision, R7-handler. Cell = pass / FAIL / n/a. If an auditor said fail but no finding survived verification, mark the cell "fail?(unconfirmed)".
4. Confirmed findings grouped by rule, each: operationId, claim, evidence file:line (use the verifier's evidence when more precise).
5. Filtered count, failed operation list, unverified-by-cap list.
Data (JSON):
${payload}`,
  { phase: 'Report', label: 'write-audit.md' })

return { ops: ops.length, audited: audited.length, confirmed: audited.reduce((n, a) => n + a.confirmed.length, 0), filtered, failed, unverifiedCap: unverifiedCap.length, agentsUsed: used }
