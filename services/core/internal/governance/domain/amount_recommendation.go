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
	slices.SortStableFunc(sorted, newestDocumentFirst)

	best := recommendedIndex(sorted)
	if best < 0 {
		return AmountRecommendation{}, false
	}
	winner := sorted[best]
	return AmountRecommendation{
		Recommended:  winner,
		Amount:       winner.amountWithVAT(),
		Alternatives: slices.Delete(sorted, best, best+1),
	}, true
}

// newestDocumentFirst orders candidates by the date printed on the document, newest first, ties by
// document id; the upload time plays no part.
func newestDocumentFirst(a, b AmountCandidate) int {
	if c := b.DocumentDate.Compare(a.DocumentDate); c != 0 {
		return c
	}
	return strings.Compare(a.DocumentID, b.DocumentID)
}

// recommendedIndex returns the index in sorted (newest first) of the first full-cost invoice, or,
// with none, of the first full-cost estimate or contract; -1 when no candidate states the full cost.
func recommendedIndex(sorted []AmountCandidate) int {
	if i := slices.IndexFunc(sorted, func(c AmountCandidate) bool {
		return c.FullCost && c.Kind == AmountDocInvoice
	}); i >= 0 {
		return i
	}
	return slices.IndexFunc(sorted, func(c AmountCandidate) bool { return c.FullCost })
}

// amountWithVAT is the amount in kopiyky with VAT included where the contractor pays it.
func (c AmountCandidate) amountWithVAT() int64 {
	if c.VATIncluded || !c.ContractorVATPayer {
		return c.Kopiyky
	}
	return c.Kopiyky + c.VATKopiyky
}
