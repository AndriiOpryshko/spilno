package app

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
)

const exampleToken = "1//refresh-token-example"

// testKey returns a 32-byte key; different seeds give different keys.
func testKey(seed byte) []byte {
	key := make([]byte, 32)
	for i := range key {
		key[i] = seed + byte(i)
	}
	return key
}

func mustCipher(t *testing.T, keyID string, key []byte) *TokenCipher {
	t.Helper()
	c, err := NewTokenCipher(keyID, key)
	if err != nil {
		t.Fatalf("NewTokenCipher(%q, %d-byte key): unexpected error %v", keyID, len(key), err)
	}
	if c == nil {
		t.Fatalf("NewTokenCipher(%q, %d-byte key) = nil cipher with nil error", keyID, len(key))
	}
	return c
}

func mustSeal(t *testing.T, c *TokenCipher, token string) SealedToken {
	t.Helper()
	s, err := c.Seal(token)
	if err != nil {
		t.Fatalf("Seal: unexpected error %v", err)
	}
	return s
}

// openNoPanic calls Open and turns a panic into an error, so one malformed input
// cannot abort the whole package run.
func openNoPanic(c *TokenCipher, s SealedToken) (token string, err error) {
	defer func() {
		if r := recover(); r != nil {
			token, err = "", fmt.Errorf("Open panicked: %v", r)
		}
	}()
	return c.Open(s)
}

// AC-1: Given a cipher with key id `k1` and a 32-byte key, When the token
// `1//refresh-token-example` is sealed, Then the result has key id `k1`, a 12-byte nonce and a
// ciphertext 16 bytes longer than the token that does not contain the token's bytes, and `Open`
// returns `1//refresh-token-example`.
func TestTokenCipher_AC1_SealHidesTokenAndOpenReturnsIt(t *testing.T) {
	c := mustCipher(t, "k1", testKey(1))
	s := mustSeal(t, c, exampleToken)

	if s.KeyID != "k1" {
		t.Errorf("KeyID = %q, want %q", s.KeyID, "k1")
	}
	if len(s.Nonce) != 12 {
		t.Errorf("len(Nonce) = %d, want 12", len(s.Nonce))
	}
	if want := len(exampleToken) + 16; len(s.Ciphertext) != want {
		t.Errorf("len(Ciphertext) = %d, want %d (token + 16-byte tag)", len(s.Ciphertext), want)
	}
	if bytes.Contains(s.Ciphertext, []byte(exampleToken)) {
		t.Errorf("Ciphertext contains the token's bytes")
	}

	got, err := c.Open(s)
	if err != nil {
		t.Fatalf("Open: unexpected error %v", err)
	}
	if got != exampleToken {
		t.Fatalf("Open = %q, want %q", got, exampleToken)
	}
}

// AC-2: Given the same cipher, When the same token is sealed twice, Then the two nonces differ,
// the two ciphertexts differ, and both open to the token.
func TestTokenCipher_AC2_SealingTwiceGivesDifferentNonceAndCiphertext(t *testing.T) {
	c := mustCipher(t, "k1", testKey(1))
	first := mustSeal(t, c, exampleToken)
	second := mustSeal(t, c, exampleToken)

	if bytes.Equal(first.Nonce, second.Nonce) {
		t.Errorf("both seals used the nonce %x", first.Nonce)
	}
	if bytes.Equal(first.Ciphertext, second.Ciphertext) {
		t.Errorf("both seals produced the ciphertext %x", first.Ciphertext)
	}
	for name, s := range map[string]SealedToken{"first": first, "second": second} {
		got, err := c.Open(s)
		if err != nil {
			t.Errorf("Open(%s): unexpected error %v", name, err)
			continue
		}
		if got != exampleToken {
			t.Errorf("Open(%s) = %q, want %q", name, got, exampleToken)
		}
	}
}

// AC-3: Given a token sealed by a cipher with key id `k1`, When a cipher with key id `k1` and a
// different 32-byte key opens it, Then the error is `ErrTokenUnreadable` and the token is `""`.
func TestTokenCipher_AC3_OtherKeyUnderSameKeyIDIsUnreadable(t *testing.T) {
	sealer := mustCipher(t, "k1", testKey(1))
	s := mustSeal(t, sealer, exampleToken)

	other := mustCipher(t, "k1", testKey(101))
	got, err := other.Open(s)
	if !errors.Is(err, ErrTokenUnreadable) {
		t.Errorf("Open error = %v, want ErrTokenUnreadable", err)
	}
	if got != "" {
		t.Errorf("Open token = %q, want empty", got)
	}
}

// AC-4: Given a sealed token, When one bit of the ciphertext is flipped, one bit of the nonce is
// flipped, the nonce is cut to 11 bytes, the nonce is empty or the ciphertext is empty, Then
// `Open` returns `ErrTokenUnreadable` each time and does not panic.
func TestTokenCipher_AC4_TamperedOrMalformedInputIsUnreadable(t *testing.T) {
	c := mustCipher(t, "k1", testKey(1))
	sealed := mustSeal(t, c, exampleToken)
	if got, err := c.Open(sealed); err != nil || got != exampleToken {
		t.Fatalf("Open(untouched) = %q, %v; want %q, nil", got, err, exampleToken)
	}

	cases := []struct {
		name   string
		mutate func(s *SealedToken)
	}{
		{"one bit of the ciphertext flipped", func(s *SealedToken) { s.Ciphertext[0] ^= 0x01 }},
		{"one bit of the tag flipped", func(s *SealedToken) { s.Ciphertext[len(s.Ciphertext)-1] ^= 0x80 }},
		{"one bit of the nonce flipped", func(s *SealedToken) { s.Nonce[0] ^= 0x01 }},
		{"nonce cut to 11 bytes", func(s *SealedToken) { s.Nonce = s.Nonce[:11] }},
		{"empty nonce", func(s *SealedToken) { s.Nonce = []byte{} }},
		{"empty ciphertext", func(s *SealedToken) { s.Ciphertext = []byte{} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := SealedToken{
				Ciphertext: bytes.Clone(sealed.Ciphertext),
				Nonce:      bytes.Clone(sealed.Nonce),
				KeyID:      sealed.KeyID,
			}
			tc.mutate(&s)

			if _, err := openNoPanic(c, s); !errors.Is(err, ErrTokenUnreadable) {
				t.Errorf("Open error = %v, want ErrTokenUnreadable", err)
			}
		})
	}
}

// AC-5: Given a token sealed by a cipher with key id `k1`, When a cipher with key id `k2` and the
// same key bytes opens it, Then the error is `ErrUnknownKey`.
func TestTokenCipher_AC5_OtherKeyIDIsUnknownKey(t *testing.T) {
	sealer := mustCipher(t, "k1", testKey(1))
	s := mustSeal(t, sealer, exampleToken)

	other := mustCipher(t, "k2", testKey(1))
	if _, err := other.Open(s); !errors.Is(err, ErrUnknownKey) {
		t.Errorf("Open error = %v, want ErrUnknownKey", err)
	}
}

// AC-6: Given a key of 0, 16, 31 or 33 bytes, or a key id that is empty or 65 characters long,
// Then `NewTokenCipher` returns `ErrInvalidKey` and a nil cipher; a 64-character key id with a
// 32-byte key is accepted; and `Seal("")` on a valid cipher returns `ErrEmptyToken`.
func TestTokenCipher_AC6_RejectsInvalidKeyAndEmptyToken(t *testing.T) {
	invalid := []struct {
		name  string
		keyID string
		key   []byte
	}{
		{"0-byte key", "k1", []byte{}},
		{"16-byte key", "k1", make([]byte, 16)},
		{"31-byte key", "k1", make([]byte, 31)},
		{"33-byte key", "k1", make([]byte, 33)},
		{"empty key id", "", testKey(1)},
		{"65-character key id", strings.Repeat("k", 65), testKey(1)},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			c, err := NewTokenCipher(tc.keyID, tc.key)
			if !errors.Is(err, ErrInvalidKey) {
				t.Errorf("NewTokenCipher error = %v, want ErrInvalidKey", err)
			}
			if c != nil {
				t.Errorf("NewTokenCipher returned a non-nil cipher")
			}
		})
	}

	t.Run("64-character key id with a 32-byte key is accepted", func(t *testing.T) {
		c, err := NewTokenCipher(strings.Repeat("k", 64), testKey(1))
		if err != nil {
			t.Errorf("NewTokenCipher error = %v, want nil", err)
		}
		if c == nil {
			t.Errorf("NewTokenCipher returned a nil cipher")
		}
	})

	t.Run("empty token", func(t *testing.T) {
		c := mustCipher(t, "k1", testKey(1))
		if _, err := c.Seal(""); !errors.Is(err, ErrEmptyToken) {
			t.Errorf("Seal(\"\") error = %v, want ErrEmptyToken", err)
		}
	})
}
