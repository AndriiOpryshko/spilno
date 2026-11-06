---
id: T17b
title: "Recommend the amount slot value from the case documents"
layer: "domain"
parent: T17
acs: ["AC-22"]
scope: governance
package: services/core/internal/governance/domain
test_cmd: "cd services/core && go test -count=1 ./internal/governance/domain/"
stub: services/core/internal/governance/domain/amount_recommendation.go
test_file: services/core/internal/governance/domain/amount_recommendation_test.go
estimate: "60-90 min"
status: "done"
---

# T17b — Recommend the amount slot value from the case documents

Slice of [T17](./T17-text-fragments-proposal-rules.md). PRD AC-22: when the documents of a справа state different
amounts for the same works, the system recommends the final amount billed and lists the other values with their
documents, so the член правління sees them before confirming. The recommendation is deterministic (no model).

## Interface

```go
type AmountDocKind string

const (
    AmountDocInvoice  AmountDocKind = "invoice"
    AmountDocEstimate AmountDocKind = "estimate"
    AmountDocContract AmountDocKind = "contract"
)

// AmountCandidate is one amount stated in one document version.
type AmountCandidate struct {
    DocumentID, VersionID string
    Kind                  AmountDocKind
    DocumentDate          time.Time // the date printed on the document
    UploadedAt            time.Time // when it was uploaded; never decides
    Kopiyky               int64     // the amount as stated
    VATIncluded           bool      // Kopiyky already includes VAT
    VATKopiyky            int64     // VAT stated separately when not included
    ContractorVATPayer    bool
    FullCost              bool      // states the full cost of the works (false for an advance)
}

// AmountRecommendation is the recommended value and every other candidate.
type AmountRecommendation struct {
    Recommended  AmountCandidate
    Amount       int64 // the recommended amount in kopiyky, VAT included where it applies
    Alternatives []AmountCandidate
}

// RecommendAmount picks the amount to propose; ok is false when no candidate can be recommended.
func RecommendAmount(candidates []AmountCandidate) (rec AmountRecommendation, ok bool)
```

## Rules

- Only full-cost candidates (`FullCost`) can be recommended; an advance never is.
- Invoices win over estimates and contracts. Among the eligible candidates of the winning group the latest
  `DocumentDate` wins; `UploadedAt` never decides. Equal document dates: the smaller `DocumentID` wins.
- With no full-cost invoice, the latest full-cost estimate or contract is recommended (an estimate superseded by a
  newer one loses).
- Amount: `Kopiyky` when `VATIncluded` or the contractor is not a VAT payer; otherwise `Kopiyky + VATKopiyky`.
- Alternatives: every other candidate (advances included), latest `DocumentDate` first, ties by `DocumentID`.
- The input slice is not modified.

## Acceptance criteria

- **AC-1** Given two full-cost invoices, the older-dated one uploaded last, Then the later-dated one is recommended.
- **AC-2** Given a full-cost invoice of a VAT payer stated without VAT (100 000.00 + VAT 20 000.00), Then `Amount`
  is 120 000.00 in kopiyky; the same invoice from a non-VAT payer gives 100 000.00.
- **AC-3** Given a full-cost invoice and a later advance invoice, Then the full-cost invoice is recommended and the
  advance is among the alternatives.
- **AC-4** Given no invoice, an estimate superseded by a newer estimate, and an older contract, Then the newer
  estimate is recommended.
- **AC-5** Given five candidates, Then `Alternatives` holds the other four, latest document date first, ties by
  document id, and the input slice keeps its order.
- **AC-6** Given no candidates, or only advances, Then ok is false.
