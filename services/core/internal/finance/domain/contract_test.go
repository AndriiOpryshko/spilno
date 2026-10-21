package domain

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"
)

func date(s string) time.Time {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		panic(err)
	}
	return t
}

func valid() ContractTerms {
	return ContractTerms{
		Counterparty: "Kyivteploenergo",
		AmountCents:  125_000,
		Currency:     CurrencyUAH,
		SignedAt:     date("2026-01-15"),
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name       string
		edit       func(*ContractTerms)
		wantFields []string
	}{
		{name: "valid open-ended", edit: func(*ContractTerms) {}},
		{name: "ends on the signing day", edit: func(c *ContractTerms) { e := c.SignedAt; c.EndsAt = &e }},
		{name: "200 multibyte characters", edit: func(c *ContractTerms) { c.Counterparty = strings.Repeat("ї", 200) }},
		{name: "empty counterparty", edit: func(c *ContractTerms) { c.Counterparty = "" }, wantFields: []string{FieldCounterparty}},
		{name: "201 characters", edit: func(c *ContractTerms) { c.Counterparty = strings.Repeat("x", 201) }, wantFields: []string{FieldCounterparty}},
		{name: "zero amount", edit: func(c *ContractTerms) { c.AmountCents = 0 }, wantFields: []string{FieldAmountCents}},
		{name: "negative amount", edit: func(c *ContractTerms) { c.AmountCents = -1 }, wantFields: []string{FieldAmountCents}},
		{name: "unknown currency", edit: func(c *ContractTerms) { c.Currency = "GBP" }, wantFields: []string{FieldCurrency}},
		{name: "ends before signed", edit: func(c *ContractTerms) { e := date("2026-01-14"); c.EndsAt = &e }, wantFields: []string{FieldEndsAt}},
		{
			name: "reports every broken invariant",
			edit: func(c *ContractTerms) {
				e := date("2025-12-31")
				*c = ContractTerms{Currency: "", SignedAt: c.SignedAt, EndsAt: &e}
			},
			wantFields: []string{FieldCounterparty, FieldAmountCents, FieldCurrency, FieldEndsAt},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terms := valid()
			tt.edit(&terms)
			err := terms.Validate()
			if tt.wantFields == nil {
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
