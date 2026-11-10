package google

import (
	"context"
	"errors"
	"net/url"
	"slices"
	"testing"
)

const (
	devConsentReturn  = "http://localhost:5180/storage/callback"
	devConsentDefault = "storage@example.test"
)

// The development consent leaves for the web callback at once, with a code the web hands back.
func TestDevConsentReturnsStraightToTheCallbackWithACode(t *testing.T) { // AC-27
	c := DevConsent{ReturnURL: devConsentReturn, Account: devConsentDefault}
	got, err := url.Parse(c.AuthURL("state-1"))
	if err != nil {
		t.Fatalf("AuthURL is not an address: %v", err)
	}
	if base := got.Scheme + "://" + got.Host + got.Path; base != devConsentReturn {
		t.Errorf("AuthURL leaves for %q, want %q", base, devConsentReturn)
	}
	if q := got.Query(); q.Get("state") != "state-1" || q.Get("code") != devConsentDefault {
		t.Errorf("AuthURL query = %v, want state=state-1 and code=%s", q, devConsentDefault)
	}
}

// Completing yields the account address the code names, so a spec can choose it (AC-28b).
func TestDevConsentCompletionYieldsTheAccountOfTheCode(t *testing.T) { // AC-27, AC-28b
	grant, err := DevConsent{ReturnURL: devConsentReturn}.Complete(context.Background(), "  Osbb-Storage@Example.Test ")
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if grant.AccountEmail != "osbb-storage@example.test" {
		t.Errorf("AccountEmail = %q, want the code, trimmed and lower-cased", grant.AccountEmail)
	}
	if grant.RefreshToken == "" || grant.DriveRootFolderID == "" {
		t.Errorf("grant = %+v, want a refresh token and a root folder", grant)
	}
	for _, scope := range []string{ScopeDriveFile, ScopeDocuments} {
		if !slices.Contains(grant.Scopes, scope) {
			t.Errorf("scopes %v lack %s", grant.Scopes, scope)
		}
	}
}

func TestDevConsentRefusesAnEmptyCode(t *testing.T) {
	_, err := DevConsent{ReturnURL: devConsentReturn}.Complete(context.Background(), " ")
	if !errors.Is(err, ErrCodeExchangeFailed) {
		t.Errorf("Complete(empty) error = %v, want ErrCodeExchangeFailed", err)
	}
}
