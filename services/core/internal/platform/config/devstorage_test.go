package config

import "testing"

// The fake storage consent is off in the production configuration and on only when asked for.
func TestProductionConfigHasNoDevStorageConsent(t *testing.T) { // AC-27
	t.Setenv("SPILNO_DEV_STORAGE", "")
	if Load().DevStorage {
		t.Error("DevStorage is on without being asked for")
	}
	t.Setenv("SPILNO_DEV_STORAGE", "true")
	if !Load().DevStorage {
		t.Error("DevStorage is off with SPILNO_DEV_STORAGE=true")
	}
}
