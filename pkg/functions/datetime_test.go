package functions

import (
	"testing"
	"time"

	"github.com/zclconf/go-cty/cty"
)

func TestTimestampFunc(t *testing.T) {
	result, err := TimestampFunc.Call([]cty.Value{})
	if err != nil {
		t.Fatalf("timestamp() error: %v", err)
	}

	// Parse the RFC3339 timestamp
	_, err = time.Parse(time.RFC3339, result.AsString())
	if err != nil {
		t.Fatalf("timestamp() returned invalid RFC3339: %v", err)
	}
}

func TestFormatDateFunc(t *testing.T) {
	tests := []struct {
		name      string
		format    string
		timestamp string
		wantErr   bool
	}{
		{
			name:      "simple date",
			format:    "YYYY-MM-DD",
			timestamp: "2024-01-15T10:30:00Z",
		},
		{
			name:      "with time",
			format:    "YYYY-MM-DD HH:mm:ss",
			timestamp: "2024-01-15T10:30:45Z",
		},
		{
			name:      "month name",
			format:    "MMM DD, YYYY",
			timestamp: "2024-01-15T10:30:00Z",
		},
		{
			name:      "invalid timestamp",
			format:    "YYYY-MM-DD",
			timestamp: "not-a-timestamp",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatDateFunc.Call([]cty.Value{
				cty.StringVal(tt.format),
				cty.StringVal(tt.timestamp),
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

			// Just verify it returns something
			if result.AsString() == "" {
				t.Error("formatdate() returned empty string")
			}
		})
	}
}

func TestTimeAddFunc(t *testing.T) {
	tests := []struct {
		name      string
		timestamp string
		duration  string
		wantErr   bool
	}{
		{
			name:      "add hours",
			timestamp: "2024-01-15T10:00:00Z",
			duration:  "2h",
		},
		{
			name:      "add minutes",
			timestamp: "2024-01-15T10:00:00Z",
			duration:  "30m",
		},
		{
			name:      "add days",
			timestamp: "2024-01-15T10:00:00Z",
			duration:  "24h",
		},
		{
			name:      "invalid timestamp",
			timestamp: "invalid",
			duration:  "1h",
			wantErr:   true,
		},
		{
			name:      "invalid duration",
			timestamp: "2024-01-15T10:00:00Z",
			duration:  "invalid",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TimeAddFunc.Call([]cty.Value{
				cty.StringVal(tt.timestamp),
				cty.StringVal(tt.duration),
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

			// Verify result is valid RFC3339
			_, err = time.Parse(time.RFC3339, result.AsString())
			if err != nil {
				t.Errorf("timeadd() returned invalid RFC3339: %v", err)
			}
		})
	}
}

func TestDayOfWeekFunc(t *testing.T) {
	result, err := DayOfWeekFunc.Call([]cty.Value{})
	if err != nil {
		t.Fatalf("day_of_week() error: %v", err)
	}

	dayName := result.AsString()

	// Should be a valid day name
	validDays := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	valid := false
	for _, day := range validDays {
		if dayName == day {
			valid = true
			break
		}
	}

	if !valid {
		t.Errorf("day_of_week() returned invalid day: %s", dayName)
	}
}

func TestFormatDateInvalidFormat(t *testing.T) {
	// Test with format string that has no recognized patterns
	result, err := FormatDateFunc.Call([]cty.Value{
		cty.StringVal("XYZ"),
		cty.StringVal("2024-01-15T10:00:00Z"),
	})

	if err != nil {
		t.Fatalf("formatdate() error: %v", err)
	}

	// Should return the format string unchanged if no patterns matched
	if result.AsString() != "XYZ" {
		t.Errorf("formatdate() with no patterns = %s, want XYZ", result.AsString())
	}
}

func TestFormatDateAllPatterns(t *testing.T) {
	timestamp := "2024-01-15T10:30:45Z"

	// Test that various patterns don't error
	patterns := []string{"YYYY", "MM", "DD", "YYYY-MM-DD"}

	for _, pattern := range patterns {
		t.Run(pattern, func(t *testing.T) {
			result, err := FormatDateFunc.Call([]cty.Value{
				cty.StringVal(pattern),
				cty.StringVal(timestamp),
			})

			if err != nil {
				t.Fatalf("formatdate() error: %v", err)
			}

			if result.AsString() == "" {
				t.Error("formatdate() returned empty string")
			}
		})
	}
}
