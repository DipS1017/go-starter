package utils

import (
	"strings"

	"github.com/google/uuid"
)

func SplitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// StringToPtr returns a pointer to the given string, or nil if the string is empty.
func StringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// StringSliceToNullStringArray is now just an identity function for []string.
func StringSliceToNullStringArray(arr []string) []string {
	return arr
}

// StringSliceToUUIDArray converts []string to []uuid.UUID, skipping invalid UUIDs.
func StringSliceToUUIDArray(arr []string) []uuid.UUID {
	var result []uuid.UUID
	for _, s := range arr {
		id, err := uuid.Parse(s)
		if err == nil {
			result = append(result, id)
		}
	}
	return result
}

// ParseCSV Helper func to parse comma-separated query params into string slices
func ParseCSV(param string) []string {
	if param == "" {
		return nil
	}
	parts := strings.Split(param, ",")
	var results []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			results = append(results, part)
		}
	}
	return results
}
