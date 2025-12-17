package utils

import (
	"testing"

	"github.com/google/uuid"
)

func TestIsValidUUID(t *testing.T) {
	valid := uuid.NewString()
	if !IsValidUUID(valid) {
		t.Errorf("IsValidUUID(%q) = false; want true", valid)
	}
	invalids := []string{"", "not-uuid"}
	for _, s := range invalids {
		if IsValidUUID(s) {
			t.Errorf("IsValidUUID(%q) = true; want false", s)
		}
	}
}

func TestArrStringToArrUUID(t *testing.T) {
	u1 := uuid.NewString()
	u2 := uuid.NewString()
	input := []string{u1, u2}
	got, err := ArrStringToArrUUID(input)
	if err != nil {
		t.Errorf("ArrStringToArrUUID(%v) unexpected error: %v", input, err)
	}
	if len(got) != len(input) {
		t.Errorf("ArrStringToArrUUID(%v) len = %d; want %d", input, len(got), len(input))
	}
	for i, s := range input {
		if got[i].String() != s {
			t.Errorf("ArrStringToArrUUID: got[%d]=%q; want %q", i, got[i].String(), s)
		}
	}

	// error path
	bad := []string{u1, "not-a-uuid"}
	_, err = ArrStringToArrUUID(bad)
	if err == nil {
		t.Errorf("ArrStringToArrUUID(%v) expected error, got nil", bad)
	}
}
