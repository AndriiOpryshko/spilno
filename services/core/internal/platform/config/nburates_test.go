package config

import "testing"

// AC-10: the rate source endpoint comes from NBU_RATES_MCP_URL and has no default in code.
func TestNBURatesMCPURLComesFromTheEnvironment(t *testing.T) {
	t.Setenv("NBU_RATES_MCP_URL", "")
	if got := Load().NBURatesMCPURL; got != "" {
		t.Errorf("NBURatesMCPURL unset = %q, want empty", got)
	}
	t.Setenv("NBU_RATES_MCP_URL", "http://127.0.0.1:3336/mcp")
	if got := Load().NBURatesMCPURL; got != "http://127.0.0.1:3336/mcp" {
		t.Errorf("NBURatesMCPURL = %q, want http://127.0.0.1:3336/mcp", got)
	}
}
