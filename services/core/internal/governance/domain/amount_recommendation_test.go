package domain

import (
	"slices"
	"testing"
)

// fullInvoice is a full-cost invoice with VAT included, dated on docDay of March 2026.
func fullInvoice(id string, docDay int, kopiyky int64) AmountCandidate {
	return AmountCandidate{
		DocumentID: id, VersionID: id + "-v1", Kind: AmountDocInvoice,
		DocumentDate: day(2026, 3, docDay), UploadedAt: day(2026, 4, 1),
		Kopiyky: kopiyky, VATIncluded: true, ContractorVATPayer: true, FullCost: true,
	}
}

func ids(cs []AmountCandidate) []string {
	out := make([]string, len(cs))
	for i, c := range cs {
		out[i] = c.DocumentID
	}
	return out
}

// AC-1: of two full-cost invoices the later-dated wins, even when the older one was uploaded last.
func TestRecommendAmount_AC1_LaterDocumentDateWinsOverLaterUpload(t *testing.T) {
	older := fullInvoice("inv-older", 5, 150_000_00)
	older.UploadedAt = day(2026, 4, 20)
	newer := fullInvoice("inv-newer", 10, 180_000_00)
	newer.UploadedAt = day(2026, 4, 2)

	rec, ok := RecommendAmount([]AmountCandidate{newer, older})
	if !ok || rec.Recommended.DocumentID != "inv-newer" || rec.Amount != 180_000_00 {
		t.Fatalf("got %q amount %d ok %v, want inv-newer amount 18000000 ok true",
			rec.Recommended.DocumentID, rec.Amount, ok)
	}
	if got := ids(rec.Alternatives); !slices.Equal(got, []string{"inv-older"}) {
		t.Fatalf("alternatives = %v, want [inv-older]", got)
	}
}

// AC-2: a VAT payer's invoice stated without VAT adds the stated VAT; a non-VAT payer's does not.
func TestRecommendAmount_AC2_AddsSeparateVATOnlyForAVATPayer(t *testing.T) {
	inv := fullInvoice("inv-1", 5, 100_000_00)
	inv.VATIncluded = false
	inv.VATKopiyky = 20_000_00

	for _, tc := range []struct {
		name     string
		vatPayer bool
		want     int64
	}{
		{"vat payer", true, 120_000_00},
		{"non-vat payer", false, 100_000_00},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := inv
			c.ContractorVATPayer = tc.vatPayer
			rec, ok := RecommendAmount([]AmountCandidate{c})
			if !ok || rec.Amount != tc.want {
				t.Fatalf("amount %d ok %v, want %d ok true", rec.Amount, ok, tc.want)
			}
		})
	}
}

// AC-3: a later advance invoice never wins over a full-cost invoice; it stays among the alternatives.
func TestRecommendAmount_AC3_AdvanceIsNeverRecommended(t *testing.T) {
	full := fullInvoice("inv-full", 5, 150_000_00)
	advance := fullInvoice("inv-advance", 20, 50_000_00)
	advance.FullCost = false

	rec, ok := RecommendAmount([]AmountCandidate{advance, full})
	if !ok || rec.Recommended.DocumentID != "inv-full" || rec.Amount != 150_000_00 {
		t.Fatalf("got %q amount %d ok %v, want inv-full amount 15000000 ok true",
			rec.Recommended.DocumentID, rec.Amount, ok)
	}
	if got := ids(rec.Alternatives); !slices.Equal(got, []string{"inv-advance"}) {
		t.Fatalf("alternatives = %v, want [inv-advance]", got)
	}
}

// AC-4: with no invoice, the newest full-cost estimate wins over an older estimate and an older contract.
func TestRecommendAmount_AC4_NewestEstimateWinsWithoutInvoice(t *testing.T) {
	contract := fullInvoice("contract", 1, 140_000_00)
	contract.Kind = AmountDocContract
	oldEst := fullInvoice("est-old", 3, 150_000_00)
	oldEst.Kind = AmountDocEstimate
	newEst := fullInvoice("est-new", 9, 165_000_00)
	newEst.Kind = AmountDocEstimate

	rec, ok := RecommendAmount([]AmountCandidate{oldEst, contract, newEst})
	if !ok || rec.Recommended.DocumentID != "est-new" || rec.Amount != 165_000_00 {
		t.Fatalf("got %q amount %d ok %v, want est-new amount 16500000 ok true",
			rec.Recommended.DocumentID, rec.Amount, ok)
	}
	if got := ids(rec.Alternatives); !slices.Equal(got, []string{"est-old", "contract"}) {
		t.Fatalf("alternatives = %v, want [est-old contract]", got)
	}
}

// AC-5: alternatives are the other four, latest document date first, ties by document id; input untouched.
func TestRecommendAmount_AC5_AlternativesOrderedAndInputKept(t *testing.T) {
	winner := fullInvoice("inv-win", 25, 200_000_00)
	advance := fullInvoice("inv-adv", 28, 40_000_00)
	advance.FullCost = false
	estB := fullInvoice("est-b", 12, 190_000_00)
	estB.Kind = AmountDocEstimate
	estA := fullInvoice("est-a", 12, 185_000_00)
	estA.Kind = AmountDocEstimate
	contract := fullInvoice("contract", 2, 170_000_00)
	contract.Kind = AmountDocContract

	in := []AmountCandidate{estB, contract, winner, advance, estA}
	before := slices.Clone(in)

	rec, ok := RecommendAmount(in)
	if !ok || rec.Recommended.DocumentID != "inv-win" {
		t.Fatalf("got %q ok %v, want inv-win ok true", rec.Recommended.DocumentID, ok)
	}
	if got, want := ids(rec.Alternatives), []string{"inv-adv", "est-a", "est-b", "contract"}; !slices.Equal(got, want) {
		t.Fatalf("alternatives = %v, want %v", got, want)
	}
	if !slices.Equal(in, before) {
		t.Fatalf("input changed: %v, want %v", ids(in), ids(before))
	}
}

// AC-6: nothing to recommend without candidates or with advances only.
func TestRecommendAmount_AC6_NoRecommendationWithoutAFullCostCandidate(t *testing.T) {
	adv := fullInvoice("inv-adv", 5, 50_000_00)
	adv.FullCost = false
	for name, in := range map[string][]AmountCandidate{
		"none":          nil,
		"advances only": {adv},
	} {
		t.Run(name, func(t *testing.T) {
			if _, ok := RecommendAmount(in); ok {
				t.Fatal("ok = true, want false")
			}
		})
	}
}
