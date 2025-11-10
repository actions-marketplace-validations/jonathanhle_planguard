package functions

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestAnyTrueFunc(t *testing.T) {
	tests := []struct {
		name     string
		input    cty.Value
		expected bool
	}{
		{
			name:     "all true",
			input:    cty.ListVal([]cty.Value{cty.True, cty.True, cty.True}),
			expected: true,
		},
		{
			name:     "some true",
			input:    cty.ListVal([]cty.Value{cty.False, cty.True, cty.False}),
			expected: true,
		},
		{
			name:     "all false",
			input:    cty.ListVal([]cty.Value{cty.False, cty.False, cty.False}),
			expected: false,
		},
		{
			name:     "empty list",
			input:    cty.ListValEmpty(cty.Bool),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := AnyTrueFunc.Call([]cty.Value{tt.input})
			if err != nil {
				t.Fatalf("AnyTrueFunc.Call() error = %v", err)
			}
			if result.True() != tt.expected {
				t.Errorf("AnyTrueFunc(%v) = %v, want %v", tt.input, result.True(), tt.expected)
			}
		})
	}
}

func TestAllTrueFunc(t *testing.T) {
	tests := []struct {
		name     string
		input    cty.Value
		expected bool
	}{
		{
			name:     "all true",
			input:    cty.ListVal([]cty.Value{cty.True, cty.True, cty.True}),
			expected: true,
		},
		{
			name:     "some false",
			input:    cty.ListVal([]cty.Value{cty.True, cty.False, cty.True}),
			expected: false,
		},
		{
			name:     "all false",
			input:    cty.ListVal([]cty.Value{cty.False, cty.False, cty.False}),
			expected: false,
		},
		{
			name:     "empty list",
			input:    cty.ListValEmpty(cty.Bool),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := AllTrueFunc.Call([]cty.Value{tt.input})
			if err != nil {
				t.Fatalf("AllTrueFunc.Call() error = %v", err)
			}
			if result.True() != tt.expected {
				t.Errorf("AllTrueFunc(%v) = %v, want %v", tt.input, result.True(), tt.expected)
			}
		})
	}
}

func TestHasFunc(t *testing.T) {
	obj := cty.ObjectVal(map[string]cty.Value{
		"name":  cty.StringVal("test"),
		"count": cty.NumberIntVal(5),
	})

	tests := []struct {
		name      string
		object    cty.Value
		attribute string
		expected  bool
	}{
		{
			name:      "has existing attribute",
			object:    obj,
			attribute: "name",
			expected:  true,
		},
		{
			name:      "has missing attribute",
			object:    obj,
			attribute: "missing",
			expected:  false,
		},
		// Note: HasFunc returns error for null values, so we don't test that case here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HasFunc.Call([]cty.Value{tt.object, cty.StringVal(tt.attribute)})
			if err != nil {
				t.Fatalf("HasFunc.Call() error = %v", err)
			}
			if result.True() != tt.expected {
				t.Errorf("HasFunc(%v, %q) = %v, want %v", tt.object, tt.attribute, result.True(), tt.expected)
			}
		})
	}
}

func TestGlobMatchFunc(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		str      string
		expected bool
	}{
		{
			name:     "exact match",
			pattern:  "test.tf",
			str:      "test.tf",
			expected: true,
		},
		{
			name:     "wildcard match",
			pattern:  "*.tf",
			str:      "main.tf",
			expected: true,
		},
		{
			name:     "no match",
			pattern:  "*.hcl",
			str:      "main.tf",
			expected: false,
		},
		{
			name:     "directory pattern",
			pattern:  "src/**/*.go",
			str:      "src/pkg/main.go",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GlobMatchFunc.Call([]cty.Value{
				cty.StringVal(tt.pattern),
				cty.StringVal(tt.str),
			})
			if err != nil {
				t.Fatalf("GlobMatchFunc.Call() error = %v", err)
			}
			if result.True() != tt.expected {
				t.Errorf("glob_match(%q, %q) = %v, want %v", tt.pattern, tt.str, result.True(), tt.expected)
			}
		})
	}
}

func TestRegexMatchFunc(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		str      string
		expected bool
		wantErr  bool
	}{
		{
			name:     "simple match",
			pattern:  "^test$",
			str:      "test",
			expected: true,
		},
		{
			name:     "no match",
			pattern:  "^test$",
			str:      "test123",
			expected: false,
		},
		{
			name:     "contains pattern",
			pattern:  "admin",
			str:      "my-admin-role",
			expected: true,
		},
		{
			name:     "digit pattern",
			pattern:  "\\d+",
			str:      "version123",
			expected: true,
		},
		{
			name:    "invalid pattern",
			pattern: "[invalid",
			str:     "test",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RegexMatchFunc.Call([]cty.Value{
				cty.StringVal(tt.pattern),
				cty.StringVal(tt.str),
			})

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error for invalid pattern")
				}
				return
			}

			if err != nil {
				t.Fatalf("RegexMatchFunc.Call() error = %v", err)
			}
			if result.True() != tt.expected {
				t.Errorf("regex_match(%q, %q) = %v, want %v", tt.pattern, tt.str, result.True(), tt.expected)
			}
		})
	}
}
