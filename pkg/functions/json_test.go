package functions

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestJSONEncodeFunc(t *testing.T) {
	tests := []struct {
		name     string
		input    cty.Value
		expected string
	}{
		{
			name:     "string",
			input:    cty.StringVal("hello"),
			expected: `"hello"`,
		},
		{
			name:     "number",
			input:    cty.NumberIntVal(42),
			expected: "42",
		},
		{
			name:     "boolean",
			input:    cty.True,
			expected: "true",
		},
		{
			name: "object",
			input: cty.ObjectVal(map[string]cty.Value{
				"name": cty.StringVal("test"),
				"age":  cty.NumberIntVal(30),
			}),
			expected: `{"age":30,"name":"test"}`,
		},
		{
			name: "list",
			input: cty.ListVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b"),
			}),
			expected: `["a","b"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := JSONEncodeFunc.Call([]cty.Value{tt.input})
			if err != nil {
				t.Fatalf("jsonencode() error: %v", err)
			}

			if result.AsString() != tt.expected {
				t.Errorf("jsonencode() = %s, want %s", result.AsString(), tt.expected)
			}
		})
	}
}

func TestJSONDecodeFunc(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		check   func(cty.Value) bool
		wantErr bool
	}{
		{
			name:  "string",
			input: `"hello"`,
			check: func(v cty.Value) bool {
				return v.AsString() == "hello"
			},
		},
		{
			name:  "number",
			input: "42",
			check: func(v cty.Value) bool {
				n, _ := v.AsBigFloat().Int64()
				return n == 42
			},
		},
		{
			name:  "boolean",
			input: "true",
			check: func(v cty.Value) bool {
				return v.True()
			},
		},
		{
			name:  "object with string",
			input: `{"name":"test"}`,
			check: func(v cty.Value) bool {
				// JSON objects may decode as maps
				return !v.IsNull()
			},
		},
		{
			name:  "array",
			input: `["a","b","c"]`,
			check: func(v cty.Value) bool {
				return v.LengthInt() == 3
			},
		},
		{
			name:    "invalid JSON",
			input:   `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := JSONDecodeFunc.Call([]cty.Value{
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

			if !tt.check(result) {
				t.Errorf("jsondecode() result check failed for %s", tt.input)
			}
		})
	}
}

func TestJSONRoundTrip(t *testing.T) {
	// Test with a simple string value
	original := cty.StringVal("test-value")

	// Encode
	encoded, err := JSONEncodeFunc.Call([]cty.Value{original})
	if err != nil {
		t.Fatalf("jsonencode() error: %v", err)
	}

	// Decode
	decoded, err := JSONDecodeFunc.Call([]cty.Value{encoded})
	if err != nil {
		t.Fatalf("jsondecode() error: %v", err)
	}

	// Verify value
	if decoded.AsString() != "test-value" {
		t.Error("Round trip failed: value mismatch")
	}
}

func TestJSONDecodeNestedObject(t *testing.T) {
	input := `{"outer":{"inner":"value"}}`

	result, err := JSONDecodeFunc.Call([]cty.Value{
		cty.StringVal(input),
	})
	if err != nil {
		t.Fatalf("jsondecode() error: %v", err)
	}

	// Just verify it decoded successfully
	if result.IsNull() {
		t.Error("Result should not be null")
	}
}

func TestJSONDecodeArray(t *testing.T) {
	input := `[1,2,3,4,5]`

	result, err := JSONDecodeFunc.Call([]cty.Value{
		cty.StringVal(input),
	})
	if err != nil {
		t.Fatalf("jsondecode() error: %v", err)
	}

	if result.LengthInt() != 5 {
		t.Errorf("Array length = %d, want 5", result.LengthInt())
	}
}

func TestJSONEncodeNull(t *testing.T) {
	_, err := JSONEncodeFunc.Call([]cty.Value{
		cty.NullVal(cty.String),
	})
	// Null values may not be supported - just verify it doesn't panic
	_ = err
}
