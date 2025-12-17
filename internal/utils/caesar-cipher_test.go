package utils

import (
	"testing"
)

func TestApplyCaesarCipher(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		shift  int
		expect string
	}{
		{
			name:   "zero shift",
			input:  "abc",
			shift:  0,
			expect: "abc",
		},
		{
			name:   "positive wrap-around",
			input:  "xyz",
			shift:  3,
			expect: "abc",
		},
		{
			name:   "negative shift",
			input:  "abc",
			shift:  -1,
			expect: "zab",
		},
		{
			name:   "shift greater than alphabet",
			input:  "abc",
			shift:  29,
			expect: "def",
		},
		{
			name:   "mixed runes with punctuation",
			input:  "Hello, World!",
			shift:  7,
			expect: "Olssv, Dvysk!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := _applyCaesarCipher(tt.input, tt.shift)
			if got != tt.expect {
				t.Errorf("ApplyCaesarCipher(%q, %d) = %q; want %q", tt.input, tt.shift, got, tt.expect)
			}
		})
	}
}
