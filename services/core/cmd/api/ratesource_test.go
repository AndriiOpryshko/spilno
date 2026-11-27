package main

import "testing"

// The startup log names the rate source without credentials or query secrets.
func TestRateSourceForLog(t *testing.T) {
	for in, want := range map[string]string{
		"http://127.0.0.1:8090/mcp":                       "http://127.0.0.1:8090/mcp",
		"http://user:secret@127.0.0.1:8090/mcp":           "http://127.0.0.1:8090/mcp",
		"https://rates.example.test/mcp?token=secret#top": "https://rates.example.test/mcp",
		"http://[::1:bad":                                 "<unparsable>",
	} {
		if got := rateSourceForLog(in); got != want {
			t.Errorf("rateSourceForLog(%q) = %q, want %q", in, got, want)
		}
	}
}
