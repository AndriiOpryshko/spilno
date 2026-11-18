package domain

import (
	"net/http"
	"os"
	"strings"
	"testing"
)

const liveContractPath = "../../../api/openapi.yaml"

// T114: another storage account is refused while the association has files (AC-26, AC-27). The code is
// written as a string here: the registry, PRD §5.1 and both contracts must agree on it.
const storageHasFilesCode = "identity.storage_has_files"

func TestStorageHasFilesCodeIsRegisteredOnceAsAConflict(t *testing.T) { // AC-27
	var found int
	for _, def := range ErrorRegistry {
		if def.Code != storageHasFilesCode {
			continue
		}
		found++
		if def.Status != http.StatusConflict {
			t.Errorf("%s: registry status %d, want %d", def.Code, def.Status, http.StatusConflict)
		}
	}
	if found != 1 {
		t.Errorf("registry holds %s %d times, want exactly once", storageHasFilesCode, found)
	}
	if got, want := len(ErrorRegistry), 18; got != want {
		t.Errorf("registry holds %d identity codes, want %d", got, want)
	}
}

func TestStorageHasFilesCodeIsInThePRDCatalogueAndBothContracts(t *testing.T) { // AC-27
	var inPRD bool
	for _, c := range prdIdentityCodes(t) {
		inPRD = inPRD || c == storageHasFilesCode
	}
	if !inPRD {
		t.Errorf("PRD §5.1 has no row for %s", storageHasFilesCode)
	}
	for _, path := range []string{contractPath, liveContractPath} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		var listed bool
		for _, line := range strings.Split(string(raw), "\n") {
			if strings.Contains(line, `"409": [`) && strings.Contains(line, storageHasFilesCode) {
				listed = true
			}
		}
		if !listed {
			t.Errorf("%s: no x-problem-types 409 entry lists %s", path, storageHasFilesCode)
		}
		if n := strings.Count(string(raw), "code: "+storageHasFilesCode); n < 1 {
			t.Errorf("%s: no example carries code: %s", path, storageHasFilesCode)
		}
	}
}
