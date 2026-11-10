package domain

import (
	"sort"
	"strings"
)

var apostropheReplacer = strings.NewReplacer("'", "", "’", "", "ʼ", "", "`", "", "‘", "")

// levelName reduces a name to a comparable form: case, spacing, apostrophe
// form and word order are ignored. The result is never stored.
func levelName(s string) string {
	words := strings.Fields(strings.ToLower(apostropheReplacer.Replace(s)))
	sort.Strings(words)
	return strings.Join(words, " ")
}

// SamePerson reports whether two names are equal after levelling.
func SamePerson(a, b string) bool {
	la := levelName(a)
	return la != "" && la == levelName(b)
}
