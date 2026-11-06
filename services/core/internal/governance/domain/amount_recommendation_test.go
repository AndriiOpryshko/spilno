package domain

import (
	"testing"
	"time"
)

func invoice(id string, date time.Time, kopiyky int64) AmountCandidate {
	return AmountCandidate{
		DocumentID: id, VersionID: id + "-v1", Kind: AmountDocInvoice,
		DocumentDate: date, UploadedAt: date, Kopiyky: kopiyky,
		VATIncluded: true, ContractorVATPayer: true, FullCost: true,
	}
}

func ids(cs []AmountCandidate) []string {
	out := make([]string, len(cs))
	for i, c := range cs {
		out[i] = c.DocumentID
	}
	return out
}

func sameIDs(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// AC-1: Given two full-cost invoices, the older-dated one uploaded last, Then the later-dated one is recommended.
func TestRecommendAmount_AC1_LaterDocumentDateWinsOverLaterUpload(t *testing.T) {
	newer := invoice("inv-newer", day(2026, time.March, 10), 150_000)
	newer.UploadedAt = day(2026, time.March, 11)
	older := invoice("inv-older", day(2026, time.March, 1), 120_000)
	older.UploadedAt = day(2026, time.April, 1) // uploaded last

	rec, ok := RecommendAmount([]AmountCandidate{newer, older})
	if !ok {
		t.Fatal("ok = false, want true")
	}
	if rec.Recommended.DocumentID != "inv-newer" || rec.Amount != 150_000 {
		t.Fatalf("recommended %q amount %d, want inv-newer 150000", rec.Recommended.DocumentID, rec.Amount)
	}
}

// AC-2: Given a full-cost invoice of a VAT payer stated without VAT (100 000.00 + VAT 20 000.00), Then Amount
// is 120 000.00 in kopiyky; the same invoice from a non-VAT payer gives 100 000.00.
func TestRecommendAmount_AC2_AddsSeparateVATForVATPayer(t *testing.T) {
	for _, tc := range []struct {
		name     string
		vatPayer bool
		want     int64
	}{
		{"vat payer", true, 12_000_000},
		{"non-vat payer", false, 10_000_000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := invoice("inv-1", day(2026, time.March, 1), 10_000_000)
			c.VATIncluded = false
			c.VATKopiyky = 2_000_000
			c.ContractorVATPayer = tc.vatPayer

			rec, ok := RecommendAmount([]AmountCandidate{c})
			if !ok || rec.Amount != tc.want {
				t.Fatalf("ok %v amount %d, want true %d", ok, rec.Amount, tc.want)
			}
		})
	}
}

// AC-3: Given a full-cost invoice and a later advance invoice, Then the full-cost invoice is recommended and the
// advance is among the alternatives.
func TestRecommendAmount_AC3_AdvanceIsNeverRecommended(t *testing.T) {
	full := invoice("inv-full", day(2026, time.March, 1), 150_000)
	advance := invoice("inv-advance", day(2026, time.March, 20), 50_000)
	advance.FullCost = false

	rec, ok := RecommendAmount([]AmountCandidate{full, advance})
	if !ok || rec.Recommended.DocumentID != "inv-full" {
		t.Fatalf("ok %v recommended %q, want true inv-full", ok, rec.Recommended.DocumentID)
	}
	if !sameIDs(ids(rec.Alternatives), []string{"inv-advance"}) {
		t.Fatalf("alternatives = %v, want [inv-advance]", ids(rec.Alternatives))
	}
}

// AC-4: Given no invoice, an estimate superseded by a newer estimate, and an older contract, Then the newer
// estimate is recommended.
func TestRecommendAmount_AC4_NewerEstimateWinsWithoutInvoice(t *testing.T) {
	contract := invoice("contract", day(2026, time.January, 10), 100_000)
	contract.Kind = AmountDocContract
	oldEst := invoice("est-old", day(2026, time.February, 1), 140_000)
	oldEst.Kind = AmountDocEstimate
	newEst := invoice("est-new", day(2026, time.February, 15), 160_000)
	newEst.Kind = AmountDocEstimate

	rec, ok := RecommendAmount([]AmountCandidate{oldEst, contract, newEst})
	if !ok || rec.Recommended.DocumentID != "est-new" || rec.Amount != 160_000 {
		t.Fatalf("ok %v recommended %q amount %d, want true est-new 160000",
			ok, rec.Recommended.DocumentID, rec.Amount)
	}
}

// AC-5: Given five candidates, Then Alternatives holds the other four, latest document date first, ties by
// document id, and the input slice keeps its order.
func TestRecommendAmount_AC5_AlternativesOrderedAndInputUntouched(t *testing.T) {
	winner := invoice("inv-win", day(2026, time.April, 1), 200_000)
	estB := invoice("est-b", day(2026, time.March, 1), 180_000)
	estB.Kind = AmountDocEstimate
	estA := invoice("est-a", day(2026, time.March, 1), 170_000)
	estA.Kind = AmountDocEstimate
	advance := invoice("inv-adv", day(2026, time.May, 1), 50_000)
	advance.FullCost = false
	contract := invoice("contract", day(2026, time.January, 1), 150_000)
	contract.Kind = AmountDocContract

	in := []AmountCandidate{estB, contract, winner, advance, estA}
	before := ids(in)

	rec, ok := RecommendAmount(in)
	if !ok || rec.Recommended.DocumentID != "inv-win" {
		t.Fatalf("ok %v recommended %q, want true inv-win", ok, rec.Recommended.DocumentID)
	}
	want := []string{"inv-adv", "est-a", "est-b", "contract"}
	if got := ids(rec.Alternatives); !sameIDs(got, want) {
		t.Fatalf("alternatives = %v, want %v", got, want)
	}
	if got := ids(in); !sameIDs(got, before) {
		t.Fatalf("input reordered to %v, want %v", got, before)
	}
}

// AC-6: Given no candidates, or only advances, Then ok is false.
// The last case is the control: a full-cost candidate added to the advances makes ok true.
func TestRecommendAmount_AC6_NothingToRecommend(t *testing.T) {
	adv1 := invoice("adv-1", day(2026, time.March, 1), 30_000)
	adv1.FullCost = false
	adv2 := invoice("adv-2", day(2026, time.March, 5), 40_000)
	adv2.FullCost = false
	adv2.Kind = AmountDocContract

	for _, tc := range []struct {
		name   string
		in     []AmountCandidate
		wantOK bool
	}{
		{"no candidates", nil, false},
		{"only advances", []AmountCandidate{adv1, adv2}, false},
		{"advances plus a full-cost invoice", []AmountCandidate{adv1, adv2, invoice("inv-1", day(2026, time.March, 2), 90_000)}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, ok := RecommendAmount(tc.in); ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
		})
	}
}
