package domain

import (
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

const contractPath = "../../../../../docs/features/board-protocol/contracts/openapi.yaml"

// contractGovernanceCodes returns every governance.<name> problem code the
// contract advertises (`code: governance.x` in examples/schemas). Job kinds
// such as governance.propose_draft are not matched: they are not `code:` values.
func contractGovernanceCodes(t *testing.T) []string {
	t.Helper()
	raw, err := os.ReadFile(contractPath)
	if err != nil {
		t.Fatalf("read contract: %v", err)
	}
	re := regexp.MustCompile(`\bcode:\s*(governance\.[a-z][a-z0-9_]*)`)
	seen := map[string]bool{}
	for _, m := range re.FindAllStringSubmatch(string(raw), -1) {
		seen[m[1]] = true
	}
	var out []string
	for c := range seen {
		out = append(out, c)
	}
	sort.Strings(out)
	return out
}

func TestErrorRegistry_ContractCodesHaveExactlyOneEntry(t *testing.T) {
	codes := contractGovernanceCodes(t)
	if len(codes) < 18 {
		t.Fatalf("extracted only %d governance codes from the contract: %v", len(codes), codes)
	}
	count := map[string]int{}
	for _, def := range ErrorRegistry {
		count[def.Code]++
	}
	for _, c := range codes {
		if n := count[c]; n != 1 {
			t.Errorf("contract code %s has %d registry entries, want exactly 1", c, n)
		}
	}
}

func TestErrorRegistry_NoStrayOrMalformedEntries(t *testing.T) {
	inContract := map[string]bool{}
	for _, c := range contractGovernanceCodes(t) {
		inContract[c] = true
	}
	for _, def := range ErrorRegistry {
		if !strings.HasPrefix(def.Code, "governance.") {
			t.Errorf("registry code %q lacks governance. prefix", def.Code)
		}
		if !inContract[def.Code] && def.Code != CodeDriveNotConnected { // advertised by association-onboarding, not board-protocol
			t.Errorf("registry code %q is not advertised by the contract", def.Code)
		}
		if def.Status < 400 || def.Status > 599 {
			t.Errorf("%s: status %d is not a 4xx/5xx", def.Code, def.Status)
		}
	}
}

func TestErrorRegistry_DriveNotConnected(t *testing.T) {
	if CodeDriveNotConnected != "governance.drive_not_connected" {
		t.Fatalf("CodeDriveNotConnected = %q", CodeDriveNotConnected)
	}
	n := 0
	for _, def := range ErrorRegistry {
		if def.Code == CodeDriveNotConnected {
			n++
			if def.Status != http.StatusConflict {
				t.Errorf("status %d, want 409", def.Status)
			}
		}
	}
	if n != 1 {
		t.Errorf("%d registry entries for %s, want exactly 1", n, CodeDriveNotConnected)
	}
}
