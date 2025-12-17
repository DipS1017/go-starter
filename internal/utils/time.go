package utils

import "time"

// TimeToPtr returns a pointer to the time if it's not zero, otherwise returns nil.
func TimeToPtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
