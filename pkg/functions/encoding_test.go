package functions

import (
	"encoding/base64"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestBase64EncodeFunc(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello",
			expected: "aGVsbG8=",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "special chars",
			input:    "hello@world!",
			expected: base64.StdEncoding.EncodeToString([]byte("hello@world!")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Base64EncodeFunc.Call([]cty.Value{
				cty.StringVal(tt.input),
			})

			if err != nil {
				t.Fatalf("base64encode() error: %v", err)
			}

			if result.AsString() != tt.expected {
				t.Errorf("base64encode(%s) = %s, want %s", tt.input, result.AsString(), tt.expected)
			}
		})
	}
}

func TestBase64DecodeFunc(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "simple string",
			input:    "aGVsbG8=",
			expected: "hello",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:    "invalid base64",
			input:   "not-valid-base64!!!",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Base64DecodeFunc.Call([]cty.Value{
				cty.StringVal(tt.input),
			})

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.AsString() != tt.expected {
				t.Errorf("base64decode(%s) = %s, want %s", tt.input, result.AsString(), tt.expected)
			}
		})
	}
}

func TestBase64GzipFunc(t *testing.T) {
	result, err := Base64GzipFunc.Call([]cty.Value{
		cty.StringVal("hello world"),
	})

	if err != nil {
		t.Fatalf("base64gzip() error: %v", err)
	}

	// Result should be valid base64
	decoded, err := base64.StdEncoding.DecodeString(result.AsString())
	if err != nil {
		t.Errorf("base64gzip() didn't return valid base64: %v", err)
	}

	// Decoded should start with gzip magic number
	if len(decoded) < 2 || decoded[0] != 0x1f || decoded[1] != 0x8b {
		t.Error("base64gzip() didn't return gzip data")
	}
}

func TestBase64GzipEmpty(t *testing.T) {
	result, err := Base64GzipFunc.Call([]cty.Value{
		cty.StringVal(""),
	})

	if err != nil {
		t.Fatalf("base64gzip() error: %v", err)
	}

	// Empty string should still produce valid gzip
	if result.AsString() == "" {
		t.Error("base64gzip(\"\") should return gzip header, not empty string")
	}
}

func TestURLEncodeFunc(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello world",
			expected: "hello+world",
		},
		{
			name:     "special chars",
			input:    "hello@world.com",
			expected: "hello%40world.com",
		},
		{
			name:     "already encoded",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "with slash",
			input:    "path/to/resource",
			expected: "path%2Fto%2Fresource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := URLEncodeFunc.Call([]cty.Value{
				cty.StringVal(tt.input),
			})

			if err != nil {
				t.Fatalf("urlencode() error: %v", err)
			}

			if result.AsString() != tt.expected {
				t.Errorf("urlencode(%s) = %s, want %s", tt.input, result.AsString(), tt.expected)
			}
		})
	}
}

func TestBase64RoundTrip(t *testing.T) {
	original := "Hello, World! 🌍"

	// Encode
	encoded, err := Base64EncodeFunc.Call([]cty.Value{
		cty.StringVal(original),
	})
	if err != nil {
		t.Fatalf("base64encode() error: %v", err)
	}

	// Decode
	decoded, err := Base64DecodeFunc.Call([]cty.Value{
		encoded,
	})
	if err != nil {
		t.Fatalf("base64decode() error: %v", err)
	}

	if decoded.AsString() != original {
		t.Errorf("Round trip failed: got %s, want %s", decoded.AsString(), original)
	}
}
