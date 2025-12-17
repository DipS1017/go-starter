package utils

import (
	"testing"
)

// TestGenerateOTP checks if the OTP is 6 digits, numeric, and unique across multiple generations.
func TestGenerateOTP(t *testing.T) {
	const expectedLength = 6
	const iterations = 100

	seen := make(map[string]bool)

	for i := 0; i < iterations; i++ {
		otp, err := GenerateOTP()
		if err != nil {
			t.Fatalf("GenerateOTP returned error: %v", err)
		}

		// Check length
		if len(otp) != expectedLength {
			t.Errorf("expected OTP length %d, got %d (OTP: %s)", expectedLength, len(otp), otp)
		}

		// Check if all characters are digits
		for _, r := range otp {
			if r < '0' || r > '9' {
				t.Errorf("OTP contains non-digit character: %q in OTP: %s", r, otp)
			}
		}
		// Check for uniqueness
		if seen[otp] {
			t.Errorf("Duplicate OTP generated: %s", otp)
		}
		seen[otp] = true
	}
}
