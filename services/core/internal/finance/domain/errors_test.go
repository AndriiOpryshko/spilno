package domain

import (
	"os"
	"regexp"
	"strconv"
	"testing"
)

const (
	prdPath      = "../../../../../docs/features/contract-uah-equivalent/PRD.md"
	contractPath = "../../../api/openapi.yaml"
)

func constants() []string {
	return []string{CodeEquivalentNotApplicable, CodeEquivalentOutOfRange, CodeRateNotPublished, CodeRateUnavailable}
}

// prdCatalogue returns code -> status from the PRD §5.1 table.
func prdCatalogue(t *testing.T) map[string]int {
	t.Helper()
	b, err := os.ReadFile(prdPath)
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile("(?m)^\\| `(finance\\.[a-z_]+)` \\| (\\d{3}) \\|")
	out := map[string]int{}
	for _, m := range re.FindAllStringSubmatch(string(b), -1) {
		n, _ := strconv.Atoi(m[2])
		out[m[1]] = n
	}
	return out
}

func TestConstantsEqualPRDCatalogue(t *testing.T) {
	cat := prdCatalogue(t)
	if len(cat) != 4 {
		t.Fatalf("PRD §5.1 catalogue has %d finance codes, want 4", len(cat))
	}
	got := map[string]bool{}
	for _, c := range constants() {
		if got[c] {
			t.Errorf("constant value %q is defined twice", c)
		}
		got[c] = true
		if _, ok := cat[c]; !ok {
			t.Errorf("constant %q is not in the PRD catalogue", c)
		}
	}
	for c := range cat {
		if !got[c] {
			t.Errorf("PRD code %q has no constant", c)
		}
	}
}

func TestRegistryHoldsEachCodeOnceWithPRDStatus(t *testing.T) {
	cat := prdCatalogue(t)
	seen := map[string]int{}
	for _, d := range ErrorRegistry {
		seen[d.Code]++
		if want, ok := cat[d.Code]; !ok || d.Status != want {
			t.Errorf("registry %q status %d, want %d", d.Code, d.Status, want)
		}
	}
	for c := range cat {
		if seen[c] != 1 {
			t.Errorf("registry lists %q %d times, want exactly once", c, seen[c])
		}
	}
}

func TestRegistryStatusesMatchContractProblemTypes(t *testing.T) {
	b, err := os.ReadFile(contractPath)
	if err != nil {
		t.Fatal(err)
	}
	line := regexp.MustCompile(`(?m)^\s+"(\d{3})": \[([^\]]*)\]`)
	code := regexp.MustCompile(`finance\.[a-z_]+`)
	want := map[string]int{}
	for _, m := range line.FindAllStringSubmatch(string(b), -1) {
		n, _ := strconv.Atoi(m[1])
		for _, c := range code.FindAllString(m[2], -1) {
			want[c] = n
		}
	}
	if len(want) != 4 {
		t.Fatalf("contract x-problem-types lists %d finance codes, want 4", len(want))
	}
	got := map[string]int{}
	for _, d := range ErrorRegistry {
		got[d.Code] = d.Status
	}
	for c, s := range want {
		if got[c] != s {
			t.Errorf("code %q: registry status %d, contract %d", c, got[c], s)
		}
	}
}

func TestDomainErrorMessageIsCode(t *testing.T) {
	e := &DomainError{Code: "finance.rate_unavailable"}
	if e.Error() != "finance.rate_unavailable" {
		t.Errorf("Error() = %q", e.Error())
	}
}
