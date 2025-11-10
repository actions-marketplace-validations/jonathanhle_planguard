package reporter

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/planguard/planguard/pkg/config"
)

func TestNewReporter(t *testing.T) {
	violations := []config.Violation{
		{RuleID: "test", RuleName: "Test", Severity: "error", Message: "msg"},
	}
	filtered := []config.FilteredViolation{}

	reporter := NewReporter(violations, filtered)
	if reporter == nil {
		t.Fatal("NewReporter returned nil")
	}

	if len(reporter.violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(reporter.violations))
	}
}

func TestFormatTextNoViolations(t *testing.T) {
	reporter := NewReporter([]config.Violation{}, []config.FilteredViolation{})
	output := reporter.FormatText()

	if !strings.Contains(output, "No violations found") {
		t.Errorf("Expected 'No violations found' message, got: %s", output)
	}
}

func TestFormatTextWithErrors(t *testing.T) {
	violations := []config.Violation{
		{
			RuleID:       "test_error",
			RuleName:     "Test Error",
			Severity:     "error",
			Message:      "Error message",
			File:         "test.tf",
			Line:         10,
			Column:       5,
			ResourceType: "aws_instance",
			ResourceName: "test",
		},
	}

	reporter := NewReporter(violations, []config.FilteredViolation{})
	output := reporter.FormatText()

	if !strings.Contains(output, "ERRORS: 1") {
		t.Error("Expected ERRORS section")
	}
	if !strings.Contains(output, "test.tf:10:5") {
		t.Error("Expected file location")
	}
	if !strings.Contains(output, "Test Error") {
		t.Error("Expected rule name")
	}
	if !strings.Contains(output, "aws_instance.test") {
		t.Error("Expected resource info")
	}
}

func TestFormatTextWithWarnings(t *testing.T) {
	violations := []config.Violation{
		{
			RuleID:       "test_warning",
			RuleName:     "Test Warning",
			Severity:     "warning",
			Message:      "Warning message",
			File:         "test.tf",
			Line:         20,
			Column:       10,
			ResourceType: "aws_s3_bucket",
			ResourceName: "bucket",
		},
	}

	reporter := NewReporter(violations, []config.FilteredViolation{})
	output := reporter.FormatText()

	if !strings.Contains(output, "WARNINGS: 1") {
		t.Error("Expected WARNINGS section")
	}
	if !strings.Contains(output, "test.tf:20:10") {
		t.Error("Expected file location")
	}
}

func TestFormatTextWithInfo(t *testing.T) {
	violations := []config.Violation{
		{
			RuleID:       "test_info",
			RuleName:     "Test Info",
			Severity:     "info",
			Message:      "Info message",
			File:         "test.tf",
			Line:         30,
			Column:       15,
			ResourceType: "aws_instance",
			ResourceName: "info",
		},
	}

	reporter := NewReporter(violations, []config.FilteredViolation{})
	output := reporter.FormatText()

	if !strings.Contains(output, "INFO: 1") {
		t.Error("Expected INFO section")
	}
}

func TestFormatTextWithRemediation(t *testing.T) {
	violations := []config.Violation{
		{
			RuleID:       "test",
			RuleName:     "Test",
			Severity:     "error",
			Message:      "Test message",
			File:         "test.tf",
			Line:         10,
			Column:       5,
			ResourceType: "aws_instance",
			ResourceName: "test",
			Remediation:  "Fix it by doing this...",
		},
	}

	reporter := NewReporter(violations, []config.FilteredViolation{})
	output := reporter.FormatText()

	if !strings.Contains(output, "Remediation:") {
		t.Error("Expected Remediation section")
	}
	if !strings.Contains(output, "Fix it by doing this") {
		t.Error("Expected remediation text")
	}
}

func TestFormatTextWithExceptions(t *testing.T) {
	violation := config.Violation{
		RuleID:       "test",
		RuleName:     "Test",
		Severity:     "error",
		Message:      "Test message",
		File:         "test.tf",
		Line:         10,
		Column:       5,
		ResourceType: "aws_instance",
		ResourceName: "test",
	}

	ticket := "TICKET-123"
	expires := "2025-12-31"

	filtered := []config.FilteredViolation{
		{
			Violation: violation,
			Exception: config.Exception{
				Rules:      []string{"test"},
				Reason:     "Legacy system",
				ApprovedBy: "security@example.com",
				Ticket:     &ticket,
				ExpiresAt:  &expires,
			},
		},
	}

	// Need at least one active violation for the exception section to appear
	activeViolation := config.Violation{
		RuleID:       "active",
		RuleName:     "Active",
		Severity:     "error",
		Message:      "Active violation",
		File:         "test.tf",
		Line:         20,
		Column:       1,
		ResourceType: "aws_s3_bucket",
		ResourceName: "bucket",
	}

	reporter := NewReporter([]config.Violation{activeViolation}, filtered)
	output := reporter.FormatText()

	if !strings.Contains(output, "EXCEPTED: 1") {
		t.Errorf("Expected EXCEPTED section, got: %s", output)
	}
	if !strings.Contains(output, "Legacy system") {
		t.Errorf("Expected exception reason, got: %s", output)
	}
	if !strings.Contains(output, "security@example.com") {
		t.Errorf("Expected approver, got: %s", output)
	}
	if !strings.Contains(output, "TICKET-123") {
		t.Errorf("Expected ticket, got: %s", output)
	}
	if !strings.Contains(output, "2025-12-31") {
		t.Errorf("Expected expiration date, got: %s", output)
	}
}

func TestFormatTextMixedSeverities(t *testing.T) {
	violations := []config.Violation{
		{
			RuleID:       "error1",
			RuleName:     "Error 1",
			Severity:     "error",
			Message:      "Error",
			File:         "test.tf",
			Line:         1,
			Column:       1,
			ResourceType: "aws_instance",
			ResourceName: "e1",
		},
		{
			RuleID:       "warn1",
			RuleName:     "Warning 1",
			Severity:     "warning",
			Message:      "Warning",
			File:         "test.tf",
			Line:         2,
			Column:       1,
			ResourceType: "aws_instance",
			ResourceName: "w1",
		},
		{
			RuleID:       "info1",
			RuleName:     "Info 1",
			Severity:     "info",
			Message:      "Info",
			File:         "test.tf",
			Line:         3,
			Column:       1,
			ResourceType: "aws_instance",
			ResourceName: "i1",
		},
	}

	reporter := NewReporter(violations, []config.FilteredViolation{})
	output := reporter.FormatText()

	if !strings.Contains(output, "ERRORS: 1") {
		t.Error("Expected ERRORS section")
	}
	if !strings.Contains(output, "WARNINGS: 1") {
		t.Error("Expected WARNINGS section")
	}
	if !strings.Contains(output, "INFO: 1") {
		t.Error("Expected INFO section")
	}
	if !strings.Contains(output, "Total: 3 violations") {
		t.Error("Expected total violations count")
	}
}

func TestFormatJSON(t *testing.T) {
	violations := []config.Violation{
		{
			RuleID:       "test",
			RuleName:     "Test",
			Severity:     "error",
			Message:      "Test message",
			File:         "test.tf",
			Line:         10,
			Column:       5,
			ResourceType: "aws_instance",
			ResourceName: "test",
		},
	}

	reporter := NewReporter(violations, []config.FilteredViolation{})
	output, err := reporter.FormatJSON()
	if err != nil {
		t.Fatalf("FormatJSON() error = %v", err)
	}

	// Verify it's valid JSON
	var parsed []config.Violation
	err = json.Unmarshal([]byte(output), &parsed)
	if err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	if len(parsed) != 1 {
		t.Errorf("Expected 1 violation in JSON, got %d", len(parsed))
	}

	if parsed[0].RuleID != "test" {
		t.Errorf("RuleID = %s, want test", parsed[0].RuleID)
	}
}

func TestFormatJSONEmpty(t *testing.T) {
	reporter := NewReporter([]config.Violation{}, []config.FilteredViolation{})
	output, err := reporter.FormatJSON()
	if err != nil {
		t.Fatalf("FormatJSON() error = %v", err)
	}

	// Should be valid empty JSON array
	var parsed []config.Violation
	err = json.Unmarshal([]byte(output), &parsed)
	if err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	if len(parsed) != 0 {
		t.Errorf("Expected empty array, got %d items", len(parsed))
	}
}

func TestFormatSARIF(t *testing.T) {
	violations := []config.Violation{
		{
			RuleID:       "test_rule",
			RuleName:     "Test Rule",
			Severity:     "error",
			Message:      "Test message",
			File:         "test.tf",
			Line:         10,
			Column:       5,
			ResourceType: "aws_instance",
			ResourceName: "test",
		},
	}

	reporter := NewReporter(violations, []config.FilteredViolation{})
	output, err := reporter.FormatSARIF()
	if err != nil {
		t.Fatalf("FormatSARIF() error = %v", err)
	}

	// Verify it's valid JSON
	var sarif map[string]interface{}
	err = json.Unmarshal([]byte(output), &sarif)
	if err != nil {
		t.Fatalf("Invalid SARIF JSON: %v", err)
	}

	// Check SARIF structure
	if sarif["version"] != "2.1.0" {
		t.Error("Expected SARIF version 2.1.0")
	}

	runs, ok := sarif["runs"].([]interface{})
	if !ok || len(runs) == 0 {
		t.Fatal("Expected runs array in SARIF")
	}

	run := runs[0].(map[string]interface{})
	tool := run["tool"].(map[string]interface{})
	driver := tool["driver"].(map[string]interface{})

	if driver["name"] != "Terraform Guardian" {
		t.Error("Expected driver name to be 'Terraform Guardian'")
	}

	results, ok := run["results"].([]interface{})
	if !ok || len(results) == 0 {
		t.Fatal("Expected results array in SARIF")
	}

	result := results[0].(map[string]interface{})
	if result["ruleId"] != "test_rule" {
		t.Errorf("Expected ruleId 'test_rule', got %v", result["ruleId"])
	}

	if result["level"] != "error" {
		t.Errorf("Expected level 'error', got %v", result["level"])
	}
}

func TestFormatSARIFEmpty(t *testing.T) {
	reporter := NewReporter([]config.Violation{}, []config.FilteredViolation{})
	output, err := reporter.FormatSARIF()
	if err != nil {
		t.Fatalf("FormatSARIF() error = %v", err)
	}

	// Verify it's valid SARIF JSON
	var sarif map[string]interface{}
	err = json.Unmarshal([]byte(output), &sarif)
	if err != nil {
		t.Fatalf("Invalid SARIF JSON: %v", err)
	}

	runs := sarif["runs"].([]interface{})
	run := runs[0].(map[string]interface{})

	// When there are no violations, results will be nil (not an empty array)
	results := run["results"]
	if results != nil {
		// If not nil, check it's empty
		if resultArray, ok := results.([]interface{}); ok && len(resultArray) != 0 {
			t.Errorf("Expected empty or nil results, got %d", len(resultArray))
		}
	}
}

func TestSeverityToLevel(t *testing.T) {
	reporter := NewReporter([]config.Violation{}, []config.FilteredViolation{})

	tests := []struct {
		severity string
		expected string
	}{
		{"error", "error"},
		{"warning", "warning"},
		{"info", "note"},
		{"unknown", "warning"}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			level := reporter.severityToLevel(tt.severity)
			if level != tt.expected {
				t.Errorf("severityToLevel(%s) = %s, want %s", tt.severity, level, tt.expected)
			}
		})
	}
}

func TestShouldFailError(t *testing.T) {
	violations := []config.Violation{
		{Severity: "error"},
		{Severity: "warning"},
	}

	reporter := NewReporter(violations, []config.FilteredViolation{})

	if !reporter.ShouldFail("error") {
		t.Error("ShouldFail(error) should return true when errors exist")
	}
}

func TestShouldFailWarning(t *testing.T) {
	violations := []config.Violation{
		{Severity: "warning"},
	}

	reporter := NewReporter(violations, []config.FilteredViolation{})

	if !reporter.ShouldFail("warning") {
		t.Error("ShouldFail(warning) should return true when warnings exist")
	}

	if reporter.ShouldFail("error") {
		t.Error("ShouldFail(error) should return false when only warnings exist")
	}
}

func TestShouldFailInfo(t *testing.T) {
	violations := []config.Violation{
		{Severity: "info"},
	}

	reporter := NewReporter(violations, []config.FilteredViolation{})

	if !reporter.ShouldFail("info") {
		t.Error("ShouldFail(info) should return true when any violations exist")
	}
}

func TestShouldFailNoViolations(t *testing.T) {
	reporter := NewReporter([]config.Violation{}, []config.FilteredViolation{})

	if reporter.ShouldFail("error") {
		t.Error("ShouldFail should return false when no violations")
	}
	if reporter.ShouldFail("warning") {
		t.Error("ShouldFail should return false when no violations")
	}
	if reporter.ShouldFail("info") {
		t.Error("ShouldFail should return false when no violations")
	}
}

func TestShouldFailDefault(t *testing.T) {
	violations := []config.Violation{
		{Severity: "error"},
	}

	reporter := NewReporter(violations, []config.FilteredViolation{})

	// Unknown failOn value should default to error
	if !reporter.ShouldFail("unknown") {
		t.Error("ShouldFail with unknown level should default to error behavior")
	}
}

func TestFilterBySeverity(t *testing.T) {
	violations := []config.Violation{
		{RuleID: "e1", Severity: "error"},
		{RuleID: "e2", Severity: "error"},
		{RuleID: "w1", Severity: "warning"},
		{RuleID: "i1", Severity: "info"},
	}

	reporter := NewReporter(violations, []config.FilteredViolation{})

	errors := reporter.filterBySeverity("error")
	if len(errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errors))
	}

	warnings := reporter.filterBySeverity("warning")
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warnings))
	}

	infos := reporter.filterBySeverity("info")
	if len(infos) != 1 {
		t.Errorf("Expected 1 info, got %d", len(infos))
	}
}

func TestIndent(t *testing.T) {
	text := "line1\nline2\nline3"
	indented := indent(text, 2)

	lines := strings.Split(indented, "\n")
	for i, line := range lines {
		if !strings.HasPrefix(line, "  ") {
			t.Errorf("Line %d not properly indented: %s", i, line)
		}
	}
}

func TestFormatViolationWithoutRemediation(t *testing.T) {
	violation := config.Violation{
		RuleID:       "test",
		RuleName:     "Test",
		Severity:     "error",
		Message:      "Test message",
		File:         "test.tf",
		Line:         10,
		Column:       5,
		ResourceType: "aws_instance",
		ResourceName: "test",
		Remediation:  "", // No remediation
	}

	reporter := NewReporter([]config.Violation{violation}, []config.FilteredViolation{})
	formatted := reporter.formatViolation(violation)

	if strings.Contains(formatted, "Remediation:") {
		t.Error("Should not include Remediation section when empty")
	}
}

func TestFormatFilteredViolationWithoutOptionalFields(t *testing.T) {
	violation := config.Violation{
		RuleID:       "test",
		RuleName:     "Test",
		Severity:     "error",
		Message:      "Test message",
		File:         "test.tf",
		Line:         10,
		Column:       5,
		ResourceType: "aws_instance",
		ResourceName: "test",
	}

	filtered := config.FilteredViolation{
		Violation: violation,
		Exception: config.Exception{
			Rules:      []string{"test"},
			Reason:     "Legacy",
			ApprovedBy: "admin@example.com",
			// No Ticket or ExpiresAt
		},
	}

	reporter := NewReporter([]config.Violation{}, []config.FilteredViolation{filtered})
	formatted := reporter.formatFilteredViolation(filtered)

	if strings.Contains(formatted, "Ticket:") {
		t.Error("Should not include Ticket when nil")
	}
	if strings.Contains(formatted, "Expires:") {
		t.Error("Should not include Expires when nil")
	}

	// Should still have required fields
	if !strings.Contains(formatted, "Legacy") {
		t.Error("Should include reason")
	}
	if !strings.Contains(formatted, "admin@example.com") {
		t.Error("Should include approver")
	}
}

func TestBuildSARIFRules(t *testing.T) {
	violations := []config.Violation{
		{
			RuleID:   "rule1",
			RuleName: "Rule 1",
			Severity: "error",
			Message:  "Message 1",
		},
		{
			RuleID:   "rule1", // Duplicate - should only appear once
			RuleName: "Rule 1",
			Severity: "error",
			Message:  "Message 1",
		},
		{
			RuleID:   "rule2",
			RuleName: "Rule 2",
			Severity: "warning",
			Message:  "Message 2",
		},
	}

	reporter := NewReporter(violations, []config.FilteredViolation{})
	rules := reporter.buildSARIFRules()

	// Should have 2 unique rules
	if len(rules) != 2 {
		t.Errorf("Expected 2 unique rules, got %d", len(rules))
	}
}

func TestBuildSARIFResults(t *testing.T) {
	violations := []config.Violation{
		{
			RuleID:   "test",
			Severity: "error",
			Message:  "Test",
			File:     "test.tf",
			Line:     10,
			Column:   5,
		},
	}

	reporter := NewReporter(violations, []config.FilteredViolation{})
	results := reporter.buildSARIFResults()

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result["ruleId"] != "test" {
		t.Error("RuleId not set correctly in SARIF result")
	}

	if result["level"] != "error" {
		t.Error("Level not set correctly in SARIF result")
	}

	locations, ok := result["locations"].([]map[string]interface{})
	if !ok || len(locations) == 0 {
		t.Fatal("Expected locations in SARIF result")
	}

	location := locations[0]
	physicalLocation := location["physicalLocation"].(map[string]interface{})
	region := physicalLocation["region"].(map[string]interface{})

	if region["startLine"] != 10 {
		t.Error("Start line not set correctly")
	}

	if region["startColumn"] != 5 {
		t.Error("Start column not set correctly")
	}
}
