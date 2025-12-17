package errorhandler

import (
	"net/http"
	"strings"
	"testing"
)

func TestErrorConstructors(t *testing.T) {
	type testCase struct {
		name     string
		fn       func(any) HttpError
		wantCode int
	}
	cases := []testCase{
		{"bad request", ErrorBadRequest, http.StatusBadRequest},      // Pass the function itself
		{"unauthorized", ErrorUnauthorized, http.StatusUnauthorized}, // Pass the function itself
		{"internal", ErrorInternal, http.StatusInternalServerError},  // Pass the function itself
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg := "boom"
			// he is now HttpError, not *HttpError
			he := tc.fn(msg)
			if he.Code != tc.wantCode {
				t.Errorf("%s: Code = %d; want %d", tc.name, he.Code, tc.wantCode)
			}
			if !strings.Contains(he.Error(), msg) {
				t.Errorf("%s: Error() = %q; want to contain %q", tc.name, he.Error(), msg)
			}
			if he.Caller == "" {
				t.Errorf("%s: Caller is empty; want non-empty", tc.name)
			}
		})
	}
}
