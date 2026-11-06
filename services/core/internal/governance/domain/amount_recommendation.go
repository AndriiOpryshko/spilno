package domain

import (
	"slices"
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
	slices.SortStableFunc(sorted, func(a, b AmountCandidate) int {
		if c := b.DocumentDate.Compare(a.DocumentDate); c != 0 {
			return c
		}
		if a.DocumentID < b.DocumentID {
			return -1
		}
		if a.DocumentID > b.DocumentID {
			return 1
		}
		return 0
	})

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
	amount := r.Kopiyky
	if !r.VATIncluded && r.ContractorVATPayer {
		amount += r.VATKopiyky
	}
	alternatives := make([]AmountCandidate, 0, len(sorted)-1)
	alternatives = append(alternatives, sorted[:best]...)
	alternatives = append(alternatives, sorted[best+1:]...)
	return AmountRecommendation{Recommended: r, Amount: amount, Alternatives: alternatives}, true
}
