package domain

import (
	"bytes"
	"errors"
	"testing"
)

func padded(head []byte, size int) []byte {
	b := make([]byte, size)
	copy(b, head)
	return b
}

func TestExtractFileValidate(t *testing.T) {
	pdf := []byte("%PDF-1.7\n")
	p7s := []byte{0x30, 0x82, 0x01, 0x00}
	zip := []byte{'P', 'K', 0x03, 0x04}
	tests := []struct {
		name string
		file ExtractFile
		ok   bool
	}{
		{"pdf", ExtractFile{"vytyah.pdf", padded(pdf, 1024)}, true},
		{"p7s container", ExtractFile{"vytyah.pdf.p7s", padded(p7s, 1024)}, true},
		{"asic-e container", ExtractFile{"vytyah.asice", padded(zip, 1024)}, true},
		{"pdf of exactly 5 MB", ExtractFile{"vytyah.pdf", padded(pdf, MaxExtractBytes)}, true},
		{"pdf one byte over 5 MB", ExtractFile{"vytyah.pdf", padded(pdf, MaxExtractBytes+1)}, false},
		{"plain text", ExtractFile{"vytyah.txt", []byte("hello world, not a pdf")}, false},
		{"png image", ExtractFile{"vytyah.png", padded([]byte{0x89, 'P', 'N', 'G'}, 1024)}, false},
		{"separate signature file of a bare signature", ExtractFile{"vytyah.sig", []byte("-----BEGIN SIGNATURE-----")}, false},
		{"empty", ExtractFile{"vytyah.pdf", nil}, false},
		{"pdf name but not pdf bytes", ExtractFile{"vytyah.pdf", bytes.Repeat([]byte("x"), 100)}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.file.Validate()
			if tc.ok && err != nil {
				t.Fatalf("Validate() = %v, want nil", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("Validate() = nil, want a refusal")
			}
		})
	}
}

func TestExtractFileMaxIs5MB(t *testing.T) {
	if MaxExtractBytes != 5*1024*1024 {
		t.Fatalf("MaxExtractBytes = %d", MaxExtractBytes)
	}
	var _ = errors.New
}
