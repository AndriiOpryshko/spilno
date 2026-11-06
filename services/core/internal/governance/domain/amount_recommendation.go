package domain

import (
	"slices"
	"strings"
	"time"
)

// AmountDocKind is the kind of case document an amount is stated in.
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
		return strings.Compare(a.DocumentID, b.DocumentID)
	})

	best := -1
	for i, c := range sorted {
		if !c.FullCost {
			continue
		}
		if c.Kind == AmountDocInvoice {
			best = i
			break
		}
		if best < 0 {
			best = i
		}
	}
	if best < 0 {
		return AmountRecommendation{}, false
	}

	winner := sorted[best]
	amount := winner.Kopiyky
	if !winner.VATIncluded && winner.ContractorVATPayer {
		amount += winner.VATKopiyky
	}
	return AmountRecommendation{
		Recommended:  winner,
		Amount:       amount,
		Alternatives: slices.Delete(sorted, best, best+1),
	}, true
}
