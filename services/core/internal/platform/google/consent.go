// Package google is the Google side of authorizing spilno against the
// association-owned Google account (board-protocol ADR-0005).
package google

import (
	"context"
	"errors"

	"golang.org/x/oauth2"
)

const (
	ScopeDriveFile = "https://www.googleapis.com/auth/drive.file"
	ScopeDocuments = "https://www.googleapis.com/auth/documents"

	// RootFolderName is the Drive folder spilno keeps an association's files in.
	RootFolderName = "spilno"
)

var (
	ErrCodeExchangeFailed = errors.New("google: authorization code exchange failed")
	ErrNoRefreshToken     = errors.New("google: the consent returned no refresh token")
	ErrScopeNotGranted    = errors.New("google: a required scope was not granted")
	ErrDriveUnavailable   = errors.New("google: drive request failed")
)

// Consent is one authorization of spilno against the association-owned Google account.
type Consent struct {
	OAuth    *oauth2.Config // client id and secret, endpoint, redirect URL
	DriveURL string         // Drive API base, "https://www.googleapis.com/drive/v3" in production; no trailing slash
}

// Grant is what a completed consent gives.
type Grant struct {
	AccountEmail      string // the Google account that gave the consent
	RefreshToken      string
	Scopes            []string // ScopeDriveFile, ScopeDocuments
	DriveRootFolderID string   // the spilno folder
}

// AuthURL returns the link the operator opens to give the consent.
func (c Consent) AuthURL(state string) string {
	return ""
}

// Complete exchanges the authorization code and prepares the account: reads its email, finds or creates the folder.
func (c Consent) Complete(ctx context.Context, code string) (Grant, error) {
	return Grant{}, errors.New("not implemented")
}
