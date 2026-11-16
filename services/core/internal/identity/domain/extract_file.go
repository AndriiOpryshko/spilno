package domain

import (
	"bytes"
	"errors"
)

// MaxExtractBytes is the largest extract file accepted.
const MaxExtractBytes = 5 * 1024 * 1024

// ExtractFile is the uploaded extract of the state register.
type ExtractFile struct {
	Name    string
	Content []byte
}

// Validate accepts one PDF or one signed container (.p7s or ASiC-E) of at most 5 MB.
func (e ExtractFile) Validate() error {
	switch {
	case len(e.Content) == 0:
		return errors.New("the extract file is empty")
	case len(e.Content) > MaxExtractBytes:
		return errors.New("the extract file is larger than 5 MB")
	case bytes.HasPrefix(e.Content, []byte("%PDF")),
		isDERSequence(e.Content),
		bytes.HasPrefix(e.Content, []byte("PK\x03\x04")):
		return nil
	}
	return errors.New("the extract must be a PDF or a signed container")
}

// isDERSequence reports whether b starts with a SEQUENCE tag and a long-form
// length of one to four bytes, which every signed container of this size has.
func isDERSequence(b []byte) bool {
	if len(b) < 2 || b[0] != 0x30 || b[1] < 0x81 || b[1] > 0x84 {
		return false
	}
	return len(b) >= 2+int(b[1]-0x80)
}
