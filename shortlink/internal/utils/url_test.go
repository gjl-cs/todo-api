package utils

import "testing"

func TestValidateLongURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid https URL",
			url:     "https://www.google.com",
			wantErr: false,
		},
		{
			name:    "valid http URL with path",
			url:     "http://example.com/page?id=1",
			wantErr: false,
		},
		{
			name:    "valid URL with port",
			url:     "http://localhost:8080/test",
			wantErr: false,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "missing scheme",
			url:     "example.com",
			wantErr: true,
		},
		{
			name:    "javascript scheme",
			url:     "javascript:alert(1)",
			wantErr: true,
		},
		{
			name:    "ftp scheme",
			url:     "ftp://example.com",
			wantErr: true,
		},
		{
			name:    "missing hostname",
			url:     "https:///page",
			wantErr: true,
		},
		{
			name:    "invalid port",
			url:     "https://example.com:99999",
			wantErr: true,
		},
		{
			name:    "URL contains spaces",
			url:     "https://example.com/a b",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLongURL(tt.url)

			if (err != nil) != tt.wantErr {
				t.Errorf(
					"ValidateLongURL(%q) error = %v, wantErr %v",
					tt.url,
					err,
					tt.wantErr,
				)
			}
		})
	}
}
