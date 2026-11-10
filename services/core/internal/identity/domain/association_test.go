package domain

import "testing"

func TestAssociationDisable(t *testing.T) {
	for _, reason := range []string{"", "   ", "\t\n"} {
		a := &Association{State: AssociationConnected}
		wantCode(t, a.Disable(reason), CodeReasonRequired) // AC-22
		if a.State != AssociationConnected || a.DisabledReason != "" {
			t.Errorf("reason %q: state changed to %q/%q", reason, a.State, a.DisabledReason)
		}
	}

	a := &Association{State: AssociationConnected}
	if err := a.Disable("Подвійне підключення"); err != nil { // AC-21
		t.Fatalf("Disable: %v", err)
	}
	if a.State != AssociationDisabled || a.DisabledReason != "Подвійне підключення" {
		t.Errorf("after Disable: state %q reason %q", a.State, a.DisabledReason)
	}

	wantCode(t, a.Disable("ще раз"), CodeAssociationStateUnchanged) // AC-24
	if a.DisabledReason != "Подвійне підключення" {
		t.Errorf("reason overwritten by a refused disable: %q", a.DisabledReason)
	}
	// the reason check wins over the state check only for a connected one;
	// a disabled association with an empty reason is still "unchanged".
	wantCode(t, a.Disable(""), CodeAssociationStateUnchanged)
}

func TestAssociationEnable(t *testing.T) {
	a := &Association{State: AssociationDisabled, DisabledReason: "x"}
	if err := a.Enable(); err != nil { // AC-23
		t.Fatalf("Enable: %v", err)
	}
	if a.State != AssociationConnected {
		t.Errorf("state = %q, want connected", a.State)
	}
	wantCode(t, a.Enable(), CodeAssociationStateUnchanged) // AC-24
}
