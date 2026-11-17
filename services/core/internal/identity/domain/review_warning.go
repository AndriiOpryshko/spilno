package domain

import (
	"fmt"
	"strings"
	"time"
)

type WarningKind string

const (
	WarningNameMismatch WarningKind = "name_mismatch"
	WarningExtractAge   WarningKind = "extract_age"
)

// extractMaxAgeDays is the age past which an extract raises a warning.
const extractMaxAgeDays = 30

// NameRole names one of the three values a name warning compares.
type NameRole string

const (
	NameTyped  NameRole = "typed"
	NameSigner NameRole = "signer"
	NameHead   NameRole = "head"
)

type ReviewWarning struct {
	Kind   WarningKind
	Detail string
	// Differing lists, for a name mismatch, the values that differ, in the order typed, signer, head;
	// empty (never nil) for any other kind.
	Differing []NameRole
}

type ExtractCheck struct {
	SignerName   string
	HeadFullName string
	ExtractDated time.Time
}

// ReviewWarnings returns the advisory warnings for a filing; none blocks it.
func ReviewWarnings(typedName string, check ExtractCheck, filingDate time.Time) []ReviewWarning {
	var ws []ReviewWarning
	if d, roles := differingNames(typedName, check); d != "" {
		ws = append(ws, ReviewWarning{Kind: WarningNameMismatch, Detail: d, Differing: roles})
	}
	if filingDate.Sub(check.ExtractDated) > extractMaxAgeDays*24*time.Hour {
		ws = append(ws, ReviewWarning{
			Kind:      WarningExtractAge,
			Detail:    fmt.Sprintf("extract dated %s is more than %d days old", check.ExtractDated.Format("2006-01-02"), extractMaxAgeDays),
			Differing: []NameRole{},
		})
	}
	return ws
}

// differingNames names the values that differ from the other two; when all
// three differ it names all of them. Empty means the three are one person.
func differingNames(typed string, c ExtractCheck) (string, []NameRole) {
	type labelled struct {
		role  NameRole
		label string
		value string
	}
	v := [3]labelled{{NameTyped, "typed name", typed}, {NameSigner, "extract signer", c.SignerName}, {NameHead, "head in extract", c.HeadFullName}}
	ts, th, sh := SamePerson(v[0].value, v[1].value), SamePerson(v[0].value, v[2].value), SamePerson(v[1].value, v[2].value)
	var odd []labelled
	switch {
	case ts && th:
		return "", nil
	case ts:
		odd = v[2:]
	case th:
		odd = v[1:2]
	case sh:
		odd = v[:1]
	default:
		odd = v[:]
	}
	parts := make([]string, len(odd))
	roles := make([]NameRole, len(odd))
	for i, o := range odd {
		parts[i] = fmt.Sprintf("%s %q", o.label, o.value)
		roles[i] = o.role
	}
	return "names differ: " + strings.Join(parts, ", "), roles
}
