package identity_test

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"

	fingen "github.com/AndriiOpryshko/spilno/services/core/internal/finance/adapters/http/gen"
	govgen "github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/http/gen"
	idgen "github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/http/gen"
)

const readOnlyReason = "спір щодо складу правління"

func readOnlyDisable(t *testing.T, e *guardEnv, assoc string) {
	t.Helper()
	req := func(method, path, body string) guardReply {
		return e.callBody(t, "operator", method, path, body)
	}
	if got := req("POST", "/identity/operator/associations/"+assoc+"/disabling", `{"reason":"`+readOnlyReason+`"}`); got.status != http.StatusOK {
		t.Fatalf("disableAssociation: status = %d (%s), want 200", got.status, got.body)
	}
}

func (e *guardEnv) callBody(t *testing.T, as, method, path, body string) guardReply {
	t.Helper()
	req, err := http.NewRequest(method, e.srv.URL+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if c := e.cookies[as]; c != nil {
		req.AddCookie(c)
	}
	resp, err := signInNoRedirect.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var buf strings.Builder
	_, _ = io.Copy(&buf, resp.Body)
	return guardReply{resp.StatusCode, []byte(buf.String())}
}

func readOnlyDetail(r guardReply) string {
	var p struct {
		Detail string `json:"detail"`
	}
	_ = json.Unmarshal(r.body, &p)
	return p.Detail
}

// AC-21, AC-25, AC-26: a member of a disabled association passes the guard only on the three allow-listed
// operations; every other association-scoped operation answers 403 identity.association_disabled with the reason.
func TestGuardMakesADisabledAssociationReadOnly(t *testing.T) {
	e := newGuardEnv(t, false)
	readOnlyDisable(t, e, e.a)

	idSpec, _ := idgen.GetSpec()
	govSpec, _ := govgen.GetSpec()
	finSpec, _ := fingen.GetSpec()
	param := regexp.MustCompile(`\{[^}]+\}`)
	allowed := map[string]bool{"searchRegistry": true, "getSignedProtocol": true, "downloadFile": true}
	passed := map[string]bool{}
	refusedN := 0
	for _, spec := range []struct {
		prefix string
		paths  map[string]map[string]string // path -> method -> operationId
	}{
		{prefix: "/governance/", paths: opIDs(govSpec.Paths.Map())},
		{prefix: "/finance/", paths: opIDs(finSpec.Paths.Map())},
		{prefix: "/identity/", paths: opIDs(idSpec.Paths.Map())},
	} {
		for path, methods := range spec.paths {
			if !strings.HasPrefix(path, spec.prefix) || !strings.Contains(path, "{association_id}") || strings.HasPrefix(path, "/identity/operator/") {
				continue
			}
			for m, opID := range methods {
				url := param.ReplaceAllString(strings.Replace(path, "{association_id}", e.a, 1), uuid.NewString())
				got := e.call(t, "member", m, url)
				if allowed[opID] {
					passed[opID] = true
					if got.code() == "identity.association_disabled" || got.code() == "identity.not_current_member" {
						t.Errorf("%s (%s %s) on a disabled association: code = %q, want it past the guard", opID, m, path, got.code())
					}
					continue
				}
				refusedN++
				if got.status != http.StatusForbidden || got.code() != "identity.association_disabled" {
					t.Errorf("%s (%s %s) on a disabled association: status = %d code = %q, want 403 identity.association_disabled", opID, m, path, got.status, got.code())
					continue
				}
				if !strings.Contains(readOnlyDetail(got), readOnlyReason) {
					t.Errorf("%s: detail = %q, want it to carry the reason %q", opID, readOnlyDetail(got), readOnlyReason)
				}
			}
		}
	}
	if len(passed) != 3 {
		t.Errorf("walked allow-listed operations %v, want all of searchRegistry, getSignedProtocol, downloadFile", passed)
	}
	if refusedN < 10 {
		t.Fatalf("checked %d refused operations, want every other association-scoped one", refusedN)
	}
}

// AC-07 comes first: a person off the roster of a disabled association still gets not_current_member.
func TestGuardKeepsNotCurrentMemberFirstOnADisabledAssociation(t *testing.T) {
	e := newGuardEnv(t, false)
	readOnlyDisable(t, e, e.a)
	for _, who := range []string{"nobody", "expired"} {
		got := e.call(t, who, "GET", e.meetings(e.a))
		if got.status != http.StatusForbidden || got.code() != "identity.not_current_member" {
			t.Errorf("%s under disabled A: status = %d code = %q, want 403 identity.not_current_member", who, got.status, got.code())
		}
	}
}

// AC-23, AC-26: a session opened before the disabling is refused on its first request after it,
// and served again after the enabling; the other association is untouched.
func TestGuardFollowsTheStateFromTheFirstRequest(t *testing.T) {
	e := newGuardEnv(t, false)
	if got := e.call(t, "member", "GET", e.meetings(e.a)); got.status != http.StatusOK {
		t.Fatalf("before disabling: status = %d (%s), want 200", got.status, got.body)
	}
	readOnlyDisable(t, e, e.a)
	got := e.call(t, "member", "GET", e.meetings(e.a))
	if got.status != http.StatusForbidden || got.code() != "identity.association_disabled" {
		t.Fatalf("first request after disabling: status = %d code = %q, want 403 identity.association_disabled", got.status, got.code())
	}
	if got := e.callBody(t, "operator", "POST", "/identity/operator/associations/"+e.a+"/enabling", ""); got.status != http.StatusOK {
		t.Fatalf("enableAssociation: status = %d (%s), want 200", got.status, got.body)
	}
	if got := e.call(t, "member", "GET", e.meetings(e.a)); got.status != http.StatusOK {
		t.Errorf("after enabling: status = %d code = %q, want 200", got.status, got.code())
	}
}

func opIDs(m map[string]*openapi3.PathItem) map[string]map[string]string {
	out := map[string]map[string]string{}
	for p, item := range m {
		out[p] = map[string]string{}
		for method, op := range item.Operations() {
			// the generated specs capitalise ids; the contract names them in lowerCamelCase.
			out[p][method] = strings.ToLower(op.OperationID[:1]) + op.OperationID[1:]
		}
	}
	return out
}
