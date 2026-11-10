// Package google is the Google side of authorizing spilno against the
// association-owned Google account (board-protocol ADR-0005).
package google

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"golang.org/x/oauth2"
)

const (
	ScopeDriveFile = "https://www.googleapis.com/auth/drive.file"
	ScopeDocuments = "https://www.googleapis.com/auth/documents"

	// RootFolderName is the Drive folder spilno keeps an association's files in.
	RootFolderName = "spilno"

	driveFolderMimeType = "application/vnd.google-apps.folder"
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
	cfg := *c.OAuth
	cfg.Scopes = requiredScopes()
	// Without offline access and a forced consent Google returns no refresh
	// token on a repeated consent.
	return cfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// Complete exchanges the authorization code and prepares the account: reads its email, finds or creates the folder.
func (c Consent) Complete(ctx context.Context, code string) (Grant, error) {
	// The cause is dropped on purpose: it may quote the token endpoint's answer.
	token, err := c.OAuth.Exchange(ctx, code)
	if err != nil {
		return Grant{}, ErrCodeExchangeFailed
	}
	if token.RefreshToken == "" {
		return Grant{}, ErrNoRefreshToken
	}
	// An answer without a scope field grants what was asked (RFC 6749 §5.1).
	if scope, ok := token.Extra("scope").(string); ok {
		granted := strings.Fields(scope)
		for _, required := range requiredScopes() {
			if !slices.Contains(granted, required) {
				return Grant{}, ErrScopeNotGranted
			}
		}
	}

	email, err := c.accountEmail(ctx, token.AccessToken)
	if err != nil {
		return Grant{}, err
	}
	folderID, err := c.rootFolder(ctx, token.AccessToken)
	if err != nil {
		return Grant{}, err
	}

	return Grant{
		AccountEmail:      email,
		RefreshToken:      token.RefreshToken,
		Scopes:            requiredScopes(),
		DriveRootFolderID: folderID,
	}, nil
}

func requiredScopes() []string { return []string{ScopeDriveFile, ScopeDocuments} }

// accountEmail reads the email of the account that gave the consent.
func (c Consent) accountEmail(ctx context.Context, accessToken string) (string, error) {
	var about struct {
		User struct {
			EmailAddress string `json:"emailAddress"`
		} `json:"user"`
	}
	query := url.Values{"fields": {"user(emailAddress)"}}
	status, err := c.drive(ctx, accessToken, http.MethodGet, "/about", query, nil, &about)
	if err != nil {
		return "", err
	}
	if about.User.EmailAddress == "" {
		return "", driveError(http.MethodGet, "/about", fmt.Sprintf("status %d, no email address", status))
	}
	return about.User.EmailAddress, nil
}

// rootFolder returns the id of the spilno folder: the oldest one spilno
// created earlier, or a new one.
func (c Consent) rootFolder(ctx context.Context, accessToken string) (string, error) {
	var list struct {
		Files []struct {
			ID string `json:"id"`
		} `json:"files"`
	}
	query := url.Values{
		"q":       {fmt.Sprintf("name = '%s' and mimeType = '%s' and trashed = false", RootFolderName, driveFolderMimeType)},
		"orderBy": {"createdTime"},
		"fields":  {"files(id)"},
	}
	status, err := c.drive(ctx, accessToken, http.MethodGet, "/files", query, nil, &list)
	if err != nil {
		return "", err
	}
	if len(list.Files) > 0 {
		if list.Files[0].ID == "" {
			return "", driveError(http.MethodGet, "/files", fmt.Sprintf("status %d, no folder id", status))
		}
		return list.Files[0].ID, nil
	}

	folder, err := json.Marshal(map[string]string{"name": RootFolderName, "mimeType": driveFolderMimeType})
	if err != nil {
		return "", driveError(http.MethodPost, "/files", "request not built")
	}
	var created struct {
		ID string `json:"id"`
	}
	status, err = c.drive(ctx, accessToken, http.MethodPost, "/files", url.Values{"fields": {"id"}}, folder, &created)
	if err != nil {
		return "", err
	}
	if created.ID == "" {
		return "", driveError(http.MethodPost, "/files", fmt.Sprintf("status %d, no folder id", status))
	}
	return created.ID, nil
}

// drive sends one Drive request with the access token and decodes the JSON
// answer into out. It returns the HTTP status of a 2xx answer.
func (c Consent) drive(ctx context.Context, accessToken, method, path string, query url.Values, body []byte, out any) (int, error) {
	var payload io.Reader
	if body != nil {
		payload = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.DriveURL+path+"?"+query.Encode(), payload)
	if err != nil {
		return 0, driveError(method, path, "request not built")
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, driveError(method, path, "no answer")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return 0, driveError(method, path, fmt.Sprintf("status %d", resp.StatusCode))
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return 0, driveError(method, path, fmt.Sprintf("status %d, unexpected body", resp.StatusCode))
	}
	return resp.StatusCode, nil
}

// driveError is ErrDriveUnavailable naming the request and what went wrong,
// never anything from the tokens.
func driveError(method, path, reason string) error {
	return fmt.Errorf("%w: %s %s: %s", ErrDriveUnavailable, method, path, reason)
}
