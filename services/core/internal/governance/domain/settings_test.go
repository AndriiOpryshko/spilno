package domain

import (
	"reflect"
	"strings"
	"testing"
)

func TestAssociationSettings_MissingSettings(t *testing.T) {
	tests := []struct {
		name string
		s    AssociationSettings
		want []string
	}{
		{"AC-42 everything missing, all named at once", AssociationSettings{}, []string{"/meeting_form", "/signers", "/register_start_number"}},
		{"blank strings count as missing", AssociationSettings{MeetingForm: "  ", Signers: "\t", RegisterStartNumber: 5}, []string{"/meeting_form", "/signers"}},
		{"only signers missing", AssociationSettings{MeetingForm: "in person", RegisterStartNumber: 1}, []string{"/signers"}},
		{"only register start missing", AssociationSettings{MeetingForm: "in person", Signers: "chair"}, []string{"/register_start_number"}},
		{"complete", AssociationSettings{MeetingForm: "in person", Signers: "chair", RegisterStartNumber: 12}, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.s.MissingSettings()
			if len(got) == 0 && len(tc.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("MissingSettings() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAssociationSettings_CheckComplete(t *testing.T) {
	if err := (AssociationSettings{MeetingForm: "a", Signers: "b", RegisterStartNumber: 1}).CheckComplete(); err != nil {
		t.Fatalf("CheckComplete() on complete settings = %v, want nil", err)
	}
	err := AssociationSettings{}.CheckComplete()
	got := pointersOf(t, err, "governance.settings_missing")
	want := []string{"/meeting_form", "/register_start_number", "/signers"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("named fields = %v, want %v", got, want)
	}
}

func TestAssociationSettings_Validate(t *testing.T) {
	ok := AssociationSettings{MeetingForm: "in person", Signers: "chair", RegisterStartNumber: 1}
	tests := []struct {
		name string
		s    AssociationSettings
		want []string
	}{
		{"valid", ok, nil},
		{"limits are inclusive", AssociationSettings{MeetingForm: strings.Repeat("я", 500), Signers: strings.Repeat("я", 500), RegisterStartNumber: 1}, nil},
		{"both too long", AssociationSettings{MeetingForm: strings.Repeat("я", 501), Signers: strings.Repeat("я", 501), RegisterStartNumber: 1}, []string{"/meeting_form", "/signers"}},
		{"negative register start", AssociationSettings{MeetingForm: "a", Signers: "b", RegisterStartNumber: -3}, []string{"/register_start_number"}},
		{"every invalid value at once", AssociationSettings{MeetingForm: strings.Repeat("x", 501), Signers: strings.Repeat("x", 501), RegisterStartNumber: -1}, []string{"/meeting_form", "/register_start_number", "/signers"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.s.Validate()
			if tc.want == nil {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			got := pointersOf(t, err, "governance.settings_invalid")
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("invalid fields = %v, want %v", got, tc.want)
			}
		})
	}
}
