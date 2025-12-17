package utils

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

const (
	defaultShift = 5
	alphabetSize = 26
)

// _applyCaesarCipher applies the Caesar cipher shift to a string.
// A positive shift encrypts, a negative shift decrypts.
// Non-alphabetic characters are returned unchanged.
func _applyCaesarCipher(text string, shift int) string {
	// Normalize shift to be an effective positive shift within 0-25
	// (shift % N + N) % N handles negative shifts correctly for modulo
	effectiveShift := (shift%alphabetSize + alphabetSize) % alphabetSize

	var result strings.Builder
	result.Grow(len(text)) // Pre-allocate for efficiency

	for _, charRune := range text {
		if unicode.IsLetter(charRune) {
			var base rune
			if unicode.IsUpper(charRune) {
				base = 'A'
			} else {
				base = 'a'
			}
			// Convert rune to 0-25 range, apply shift, wrap around, convert back
			shiftedRune := (charRune-base+rune(effectiveShift))%alphabetSize + base
			result.WriteRune(shiftedRune)
		} else {
			// Non-alphabetic characters are preserved
			result.WriteRune(charRune)
		}
	}
	return result.String()
}

// EncryptAndEncodeURL encrypts plain text using Caesar cipher and then URL-encodes the result.
// It accepts an optional shift value; if not provided, defaultShift is used.
func EncryptAndEncodeURL(plainText string, shiftOpt ...int) string {
	shift := defaultShift
	if len(shiftOpt) > 0 {
		shift = shiftOpt[0]
	}

	caesarEncryptedText := _applyCaesarCipher(plainText, shift)
	return url.QueryEscape(caesarEncryptedText)
}

// DecodeURLAndDecrypt URL-decodes the input string and then decrypts it using Caesar cipher.
// It accepts an optional shift value; if not provided, defaultShift is used.
// Returns the decrypted string and an error if URL decoding fails.
func DecodeURLAndDecrypt(urlSafeEncryptedText string, shiftOpt ...int) (string, error) {
	shift := defaultShift
	if len(shiftOpt) > 0 {
		shift = shiftOpt[0]
	}

	caesarEncryptedText, err := url.QueryUnescape(urlSafeEncryptedText)
	if err != nil {
		return "", fmt.Errorf("failed to URL decode: %w", err)
	}

	// For decryption, apply a negative shift
	decryptedText := _applyCaesarCipher(caesarEncryptedText, -shift)
	return decryptedText, nil
}
