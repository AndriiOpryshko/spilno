package httpapi_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
)

func decode(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not JSON: %v: %s", err, rec.Body.String())
	}
	return body
}

func TestWriteDomainError_RendersCodeAndFieldErrors(t *testing.T) {
	derr := &httpapi.DomainError{
		Code:   "governance.settings_invalid",
		Status: http.StatusUnprocessableEntity,
		Detail: "Налаштування містять помилки",
		Errors: []httpapi.FieldError{{Pointer: "/register_start_number", Detail: "must be >= 1"}},
	}
	// wrapped, as a use case would return it
	err := fmt.Errorf("save settings: %w", derr)

	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/governance/settings", nil)
	if !httpapi.WriteDomainError(rec, r, err) {
		t.Fatal("WriteDomainError returned false for a registered DomainError")
	}

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("Content-Type = %q, want application/problem+json", ct)
	}
	body := decode(t, rec)
	want := map[string]any{
		"type":   "about:blank",
		"title":  "Unprocessable Entity",
		"status": float64(422),
		"code":   "governance.settings_invalid",
		"detail": "Налаштування містять помилки",
	}
	for k, v := range want {
		if body[k] != v {
			t.Errorf("%s = %v, want %v", k, body[k], v)
		}
	}
	errs, _ := body["errors"].([]any)
	if len(errs) != 1 {
		t.Fatalf("errors = %v, want 1 entry", body["errors"])
	}
	first, _ := errs[0].(map[string]any)
	if first["pointer"] != "/register_start_number" || first["detail"] != "must be >= 1" {
		t.Errorf("errors[0] = %v", first)
	}
}

func TestWriteDomainError_IgnoresUnregisteredErrors(t *testing.T) {
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if httpapi.WriteDomainError(rec, r, errors.New("boom")) {
		t.Error("WriteDomainError returned true for a plain error")
	}
	if rec.Body.Len() != 0 {
		t.Errorf("wrote a body for a plain error: %s", rec.Body.String())
	}
}

func TestWriteProblem_ValidatorStyle422HasNoCode(t *testing.T) {
	rec := httptest.NewRecorder()
	httpapi.WriteProblem(rec, httpapi.Problem{
		Title:  "Validation failed",
		Status: http.StatusUnprocessableEntity,
		Errors: []httpapi.FieldError{{Pointer: "/body/name", Detail: "required"}},
	})
	body := decode(t, rec)
	if _, ok := body["code"]; ok {
		t.Errorf("validator 422 must not carry code, got %v", body["code"])
	}
	if body["type"] != "about:blank" {
		t.Errorf("type = %v", body["type"])
	}
}
