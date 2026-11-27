package main

import "net/url"

// rateSourceForLog is the rate-source address as the startup log shows it:
// scheme, host and path only, so credentials in the userinfo or the query
// never reach the logs.
func rateSourceForLog(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "<unparsable>"
	}
	return (&url.URL{Scheme: u.Scheme, Host: u.Host, Path: u.Path}).String()
}
