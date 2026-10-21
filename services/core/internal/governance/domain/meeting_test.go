package domain

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"
)

func mustTime(layout, s string) time.Time {
	t, err := time.Parse(layout, s)
	if err != nil {
		panic(err)
	}
	return t
}

func ptr[T any](v T) *T { return &v }

func valid() MeetingFields {
	return MeetingFields{
		Title:          "Sample title",
		Kind:           MeetingKindRegular,
		HeldAt:         mustTime(time.RFC3339, "2026-01-15T10:00:00Z"),
		EndsAt:         ptr(mustTime(time.RFC3339, "2026-01-15T12:00:00Z")),
		Location:       ptr("Sample location"),
		QuorumRequired: false,
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name       string
		edit       func(*MeetingFields)
		wantFields []string
	}{
		{name: "valid", edit: func(*MeetingFields) {}},
		{name: "title too short", edit: func(x *MeetingFields) { x.Title = strings.Repeat("x", 0) }, wantFields: []string{FieldTitle}},
		{name: "title at max in multibyte runes", edit: func(x *MeetingFields) { x.Title = strings.Repeat("ї", 200) }},
		{name: "title too long", edit: func(x *MeetingFields) { x.Title = strings.Repeat("x", 201) }, wantFields: []string{FieldTitle}},
		{name: "kind unknown", edit: func(x *MeetingFields) { x.Kind = "UNKNOWN" }, wantFields: []string{FieldKind}},
		{name: "location at max in multibyte runes", edit: func(x *MeetingFields) { x.Location = ptr(strings.Repeat("ї", 200)) }},
		{name: "no ends_at", edit: func(x *MeetingFields) { x.EndsAt = nil }},
		{name: "ends_at equal to held_at", edit: func(x *MeetingFields) { x.EndsAt = ptr(x.HeldAt) }, wantFields: []string{FieldEndsAt}},
		{name: "ends_at before held_at", edit: func(x *MeetingFields) { x.EndsAt = ptr(x.HeldAt.Add(-time.Minute)) }, wantFields: []string{FieldEndsAt}},
		{name: "location too long", edit: func(x *MeetingFields) { x.Location = ptr(strings.Repeat("x", 201)) }, wantFields: []string{FieldLocation}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := valid()
			tt.edit(&x)
			err := x.Validate()
			if len(tt.wantFields) == 0 {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			var v *ValidationError
			if !errors.As(err, &v) {
				t.Fatalf("Validate() = %v, want *ValidationError", err)
			}
			var got []string
			for _, f := range v.Fields {
				got = append(got, f.Field)
			}
			if !slices.Equal(got, tt.wantFields) {
				t.Errorf("fields = %v, want %v", got, tt.wantFields)
			}
		})
	}
}
