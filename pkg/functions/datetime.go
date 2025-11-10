package functions

import (
	"fmt"
	"time"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// TimestampFunc returns the current timestamp
var TimestampFunc = function.New(&function.Spec{
	Params: []function.Parameter{},
	Type:   function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		return cty.StringVal(time.Now().UTC().Format(time.RFC3339)), nil
	},
})

// FormatDateFunc formats a timestamp with a format string
var FormatDateFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "format", Type: cty.String},
		{Name: "timestamp", Type: cty.String},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		format := args[0].AsString()
		timestamp := args[1].AsString()

		// Parse timestamp
		t, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			return cty.NilVal, fmt.Errorf("invalid timestamp: %w", err)
		}

		// Convert format string (Terraform uses different format than Go)
		goFormat := convertTerraformDateFormat(format)
		formatted := t.Format(goFormat)

		return cty.StringVal(formatted), nil
	},
})

// TimeAddFunc adds a duration to a timestamp
var TimeAddFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "timestamp", Type: cty.String},
		{Name: "duration", Type: cty.String},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		timestamp := args[0].AsString()
		duration := args[1].AsString()

		// Parse timestamp
		t, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			return cty.NilVal, fmt.Errorf("invalid timestamp: %w", err)
		}

		// Parse duration
		d, err := time.ParseDuration(duration)
		if err != nil {
			return cty.NilVal, fmt.Errorf("invalid duration: %w", err)
		}

		result := t.Add(d)
		return cty.StringVal(result.Format(time.RFC3339)), nil
	},
})

// convertTerraformDateFormat converts Terraform date format to Go time format
// This is a simplified version - full implementation would handle all format codes
func convertTerraformDateFormat(tfFormat string) string {
	// Map common Terraform format codes to Go format codes
	replacements := map[string]string{
		"YYYY": "2006",
		"YY":   "06",
		"MM":   "01",
		"DD":   "02",
		"hh":   "15",
		"mm":   "04",
		"ss":   "05",
	}

	result := tfFormat
	for tf, go_ := range replacements {
		result = replaceAll(result, tf, go_)
	}

	return result
}

func replaceAll(s, old, new string) string {
	// Simple string replace (use strings.ReplaceAll in production)
	result := ""
	for len(s) > 0 {
		if len(s) >= len(old) && s[:len(old)] == old {
			result += new
			s = s[len(old):]
		} else {
			result += s[:1]
			s = s[1:]
		}
	}
	return result
}
