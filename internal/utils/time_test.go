package utils

import (
	"testing"
	"time"
)

func TestTimeToPtr(t *testing.T) {
	var zero time.Time
	if got := TimeToPtr(zero); got != nil {
		t.Errorf("TimeToPtr(zero) = %v; want nil", got)
	}
	nonzero := time.Now()
	got := TimeToPtr(nonzero)
	if got == nil {
		t.Fatalf("TimeToPtr(nonzero) = nil; want non-nil")
	}
	if !got.Equal(nonzero) {
		t.Errorf("TimeToPtr(%v) = %v; want times to be Equal", nonzero, *got)
	}
}
