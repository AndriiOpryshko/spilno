package google

import (
	"context"
	"net/url"
	"strings"
)

// DevConsent is the development and e2e fake of Consent; never wire it in production. Only the
// explicit configuration (SPILNO_DEV_STORAGE) may turn it on. Its authorization address returns
// straight to ReturnURL with Account as the code; completing yields the account the code names.
type DevConsent struct {
	ReturnURL string
	Account   string
}

// AuthURL returns ReturnURL carrying the code and the state.
func (c DevConsent) AuthURL(state string) string {
	u, err := url.Parse(c.ReturnURL)
	if err != nil {
		return c.ReturnURL
	}
	q := u.Query()
	q.Set("code", c.Account)
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return u.String()
}

// Complete treats the code as the address of the account that gave the consent.
func (c DevConsent) Complete(_ context.Context, code string) (Grant, error) {
	account := strings.ToLower(strings.TrimSpace(code))
	if account == "" {
		return Grant{}, ErrCodeExchangeFailed
	}
	return Grant{
		AccountEmail:      account,
		RefreshToken:      "dev-refresh-token",
		Scopes:            []string{ScopeDriveFile, ScopeDocuments},
		DriveRootFolderID: "dev-root-folder",
	}, nil
}
