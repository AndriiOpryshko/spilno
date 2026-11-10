package domain

import (
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const (
	prdPath      = "../../../../../docs/features/association-onboarding/PRD.md"
	contractPath = "../../../../../docs/features/association-onboarding/contracts/openapi.yaml"
)

// wantStatus is the status each identity.* code of PRD §5.1 is advertised with.
var wantStatus = map[string]int{
	CodeApplicationInvalid:          http.StatusUnprocessableEntity,
	CodeSignatureNotChecked:         http.StatusConflict,
	CodeExtractCodeMismatch:         http.StatusConflict,
	CodeAssociationAlreadyConnected: http.StatusConflict,
	CodeApplicationAlreadyOpen:      http.StatusConflict,
	CodeNotCurrentMember:            http.StatusForbidden,
	CodeMismatchNeedsDocument:       http.StatusConflict,
	CodeReasonRequired:              http.StatusUnprocessableEntity,
	CodeApplicationNotPending:       http.StatusConflict,
	CodeNotOperator:                 http.StatusForbidden,
	CodeOperatorOwnApplication:      http.StatusConflict,
	CodeAssociationNotConnected:     http.StatusConflict,
	CodeAssociationStateUnchanged:   http.StatusConflict,
	CodeAssociationDisabled:         http.StatusForbidden,
	CodeTermEnded:                   http.StatusConflict,
	CodeStorageAccountPersonal:      http.StatusConflict,
	CodeSessionRequired:             http.StatusUnauthorized,
}

// prdIdentityCodes returns the identity.* codes of the PRD §5.1 catalogue.
func prdIdentityCodes(t *testing.T) []string {
	t.Helper()
	raw, err := os.ReadFile(prdPath)
	if err != nil {
		t.Fatalf("read PRD: %v", err)
	}
	section := string(raw)
	i := strings.Index(section, "## 5.1 Error catalogue")
	j := strings.Index(section, "## 6. Non-functional")
	if i < 0 || j < i {
		t.Fatalf("PRD §5.1 not found")
	}
	re := regexp.MustCompile("(?m)^\\| `(identity\\.[a-z][a-z0-9_]*)` \\|")
	var out []string
	for _, m := range re.FindAllStringSubmatch(section[i:j], -1) {
		out = append(out, m[1])
	}
	sort.Strings(out)
	return out
}

func TestErrorRegistry_EqualsPRDCatalogueExactlyOnce(t *testing.T) {
	prd := prdIdentityCodes(t)
	if len(prd) < 15 {
		t.Fatalf("extracted only %d identity codes from PRD §5.1: %v", len(prd), prd)
	}
	count := map[string]int{}
	for _, def := range ErrorRegistry {
		count[def.Code]++
	}
	inPRD := map[string]bool{}
	for _, c := range prd {
		inPRD[c] = true
		if n := count[c]; n != 1 {
			t.Errorf("PRD code %s has %d registry entries, want exactly 1", c, n)
		}
	}
	for _, def := range ErrorRegistry {
		// identity.session_required predates §5.1 (board-protocol ADR-0003); the contract adds it beyond the catalogue.
		if !inPRD[def.Code] && def.Code != CodeSessionRequired {
			t.Errorf("registry code %q is not in PRD §5.1", def.Code)
		}
	}
}

func TestErrorRegistry_ConstantsMatchCatalogueStrings(t *testing.T) {
	want := map[string]string{
		"CodeApplicationInvalid":          "identity.application_invalid",
		"CodeSignatureNotChecked":         "identity.signature_not_checked",
		"CodeExtractCodeMismatch":         "identity.extract_code_mismatch",
		"CodeAssociationAlreadyConnected": "identity.association_already_connected",
		"CodeApplicationAlreadyOpen":      "identity.application_already_open",
		"CodeNotCurrentMember":            "identity.not_current_member",
		"CodeMismatchNeedsDocument":       "identity.mismatch_needs_document",
		"CodeReasonRequired":              "identity.reason_required",
		"CodeApplicationNotPending":       "identity.application_not_pending",
		"CodeNotOperator":                 "identity.not_operator",
		"CodeOperatorOwnApplication":      "identity.operator_own_application",
		"CodeAssociationNotConnected":     "identity.association_not_connected",
		"CodeAssociationStateUnchanged":   "identity.association_state_unchanged",
		"CodeAssociationDisabled":         "identity.association_disabled",
		"CodeTermEnded":                   "identity.term_ended",
		"CodeStorageAccountPersonal":      "identity.storage_account_personal",
		"CodeSessionRequired":             "identity.session_required",
	}
	got := map[string]string{
		"CodeApplicationInvalid":          CodeApplicationInvalid,
		"CodeSignatureNotChecked":         CodeSignatureNotChecked,
		"CodeExtractCodeMismatch":         CodeExtractCodeMismatch,
		"CodeAssociationAlreadyConnected": CodeAssociationAlreadyConnected,
		"CodeApplicationAlreadyOpen":      CodeApplicationAlreadyOpen,
		"CodeNotCurrentMember":            CodeNotCurrentMember,
		"CodeMismatchNeedsDocument":       CodeMismatchNeedsDocument,
		"CodeReasonRequired":              CodeReasonRequired,
		"CodeApplicationNotPending":       CodeApplicationNotPending,
		"CodeNotOperator":                 CodeNotOperator,
		"CodeOperatorOwnApplication":      CodeOperatorOwnApplication,
		"CodeAssociationNotConnected":     CodeAssociationNotConnected,
		"CodeAssociationStateUnchanged":   CodeAssociationStateUnchanged,
		"CodeAssociationDisabled":         CodeAssociationDisabled,
		"CodeTermEnded":                   CodeTermEnded,
		"CodeStorageAccountPersonal":      CodeStorageAccountPersonal,
		"CodeSessionRequired":             CodeSessionRequired,
	}
	for name, s := range want {
		if got[name] != s {
			t.Errorf("%s = %q, want %q", name, got[name], s)
		}
	}
}

func TestErrorRegistry_StatusMatchesTable(t *testing.T) {
	for _, def := range ErrorRegistry {
		want, ok := wantStatus[def.Code]
		if !ok {
			t.Errorf("registry code %s has no expected status", def.Code)
			continue
		}
		if def.Status != want {
			t.Errorf("%s: status %d, want %d", def.Code, def.Status, want)
		}
	}
	if len(ErrorRegistry) != len(wantStatus) {
		t.Errorf("registry has %d entries, want %d", len(ErrorRegistry), len(wantStatus))
	}
}

func TestErrorRegistry_StatusMatchesContractProblemTypes(t *testing.T) {
	raw, err := os.ReadFile(contractPath)
	if err != nil {
		t.Fatalf("read contract: %v", err)
	}
	re := regexp.MustCompile(`"(\d{3})":\s*\[([^\]]*)\]`)
	inContract := map[string]int{}
	for _, line := range strings.Split(string(raw), "\n") {
		m := re.FindStringSubmatch(line)
		if m == nil || !strings.Contains(m[2], "identity.") {
			continue
		}
		status, _ := strconv.Atoi(m[1])
		for _, c := range strings.Split(m[2], ",") {
			if c = strings.TrimSpace(c); strings.HasPrefix(c, "identity.") {
				inContract[c] = status
			}
		}
	}
	if len(inContract) < 10 {
		t.Fatalf("extracted only %d codes from x-problem-types: %v", len(inContract), inContract)
	}
	registry := map[string]int{}
	for _, def := range ErrorRegistry {
		registry[def.Code] = def.Status
	}
	for c, status := range inContract {
		got, ok := registry[c]
		if !ok {
			t.Errorf("contract code %s is not in the registry", c)
		} else if got != status {
			t.Errorf("%s: registry status %d, contract %d", c, got, status)
		}
	}
}

func TestCodedFieldsError_NamesEveryField(t *testing.T) {
	err := &CodedFieldsError{Code: CodeApplicationInvalid, Fields: []FieldIssue{
		{Pointer: "/edrpou", Detail: "required"}, {Pointer: "/term_start", Detail: "required"},
	}}
	if got := strings.Join(err.Pointers(), ","); got != "/edrpou,/term_start" {
		t.Errorf("Pointers = %q", got)
	}
	if !strings.HasPrefix(err.Error(), "identity.application_invalid") {
		t.Errorf("Error() = %q", err.Error())
	}
}
