package domain

import (
	"cmp"
	"slices"
	"strings"
	"time"
)

// AmountDocKind is the kind of document an amount candidate comes from.
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
	FullCost              bool // states the full cost of the works (false for an advance)
}

// AmountRecommendation is the recommended value and every other candidate.
type AmountRecommendation struct {
	Recommended  AmountCandidate
	Amount       int64 // the recommended amount in kopiyky, VAT included where it applies
	Alternatives []AmountCandidate
}

// RecommendAmount picks the amount to propose; ok is false when no candidate can be recommended.
func RecommendAmount(candidates []AmountCandidate) (rec AmountRecommendation, ok bool) {
	sorted := slices.Clone(candidates)
	slices.SortStableFunc(sorted, latestDocumentFirst)

	best := -1
	for i, c := range sorted {
		if !c.FullCost {
			continue
		}
		if best == -1 || (c.Kind == AmountDocInvoice && sorted[best].Kind != AmountDocInvoice) {
			best = i
		}
	}
	if best == -1 {
		return AmountRecommendation{}, false
	}

	r := sorted[best]
	alternatives := make([]AmountCandidate, 0, len(sorted)-1)
	alternatives = append(alternatives, sorted[:best]...)
	alternatives = append(alternatives, sorted[best+1:]...)
	return AmountRecommendation{Recommended: r, Amount: amountWithVAT(r), Alternatives: alternatives}, true
}

// latestDocumentFirst orders candidates by DocumentDate, latest first, then by DocumentID.
func latestDocumentFirst(a, b AmountCandidate) int {
	return cmp.Or(
		b.DocumentDate.Compare(a.DocumentDate),
		strings.Compare(a.DocumentID, b.DocumentID),
	)
}

// amountWithVAT is the stated amount plus the separately stated VAT when the contractor is a VAT payer.
func amountWithVAT(c AmountCandidate) int64 {
	if c.VATIncluded || !c.ContractorVATPayer {
		return c.Kopiyky
	}
	return c.Kopiyky + c.VATKopiyky
}
