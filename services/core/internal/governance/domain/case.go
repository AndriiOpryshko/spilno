package domain

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"
)

// MaxDocumentSize is the largest document an upload may carry (PRD §6): 20 MiB.
const MaxDocumentSize = 20 << 20

// ErrCaseNotFound is returned when the association has no справа with the id.
var ErrCaseNotFound = errors.New("case not found")

// ErrPreviousVersionInvalid is returned when previous_version_id is not a document of the справа.
var ErrPreviousVersionInvalid = errors.New("previous version is not a document of the case")

// ValidateCaseTitle requires 1 to MaxTitleLen characters that are not all whitespace.
func ValidateCaseTitle(title string) error {
	if n := utf8.RuneCountInString(title); n > MaxTitleLen || strings.TrimSpace(title) == "" {
		var v ValidationError
		v.add(FieldTitle, fmt.Sprintf("must be 1 to %d characters, not only whitespace", MaxTitleLen))
		return &v
	}
	return nil
}

// SniffDocumentMediaType names the media type of a document by its content, never by its name or
// declared type; ok is false unless it is PDF, JPEG or PNG.
func SniffDocumentMediaType(data []byte) (mediaType string, ok bool) {
	switch mt := http.DetectContentType(data); mt {
	case "application/pdf", "image/jpeg", "image/png":
		return mt, true
	}
	return "", false
}
