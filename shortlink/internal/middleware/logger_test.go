package middleware

import "testing"

func TestIsValidRequestID(t *testing.T) {
	tests := []struct {
		name  string
		id    string
		valid bool
	}{
		{
			name:  "valid request ID",
			id:    "request-001_test",
			valid: true,
		},
		{
			name:  "empty request ID",
			id:    "",
			valid: false,
		},
		{
			name:  "contains spaces",
			id:    "request 001",
			valid: false,
		},
		{
			name:  "contains special characters",
			id:    "request@001",
			valid: false,
		},
		{
			name:  "too long",
			id:    string(make([]byte, 65)),
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidRequestID(tt.id)

			if got != tt.valid {
				t.Errorf(
					"isValidRequestID(%q) = %v, want %v",
					tt.id,
					got,
					tt.valid,
				)
			}
		})
	}
}
