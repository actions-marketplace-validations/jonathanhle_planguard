package reporter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jonathanhle/planguard/pkg/config"
)

// Reporter handles violation reporting and formatting
type Reporter struct {
	violations         []config.Violation
	filteredViolations []config.FilteredViolation
}

// NewReporter creates a new reporter
func NewReporter(violations []config.Violation, filtered []config.FilteredViolation) *Reporter {
	return &Reporter{
		violations:         violations,
		filteredViolations: filtered,
	}
}

// FormatText formats violations as human-readable text
func (r *Reporter) FormatText() string {
	if len(r.violations) == 0 {
		return "✅ No violations found!\n"
	}

	var output strings.Builder

	// Group by severity
	errors := r.filterBySeverity("error")
	warnings := r.filterBySeverity("warning")
	infos := r.filterBySeverity("info")

	output.WriteString("🔒 Terraform Guardian Scan Results\n")
	output.WriteString(strings.Repeat("=", 50) + "\n\n")

	if len(errors) > 0 {
		output.WriteString(fmt.Sprintf("❌ ERRORS: %d\n", len(errors)))
		output.WriteString(strings.Repeat("-", 50) + "\n")
		for _, v := range errors {
			output.WriteString(r.formatViolation(v))
		}
		output.WriteString("\n")
	}

	if len(warnings) > 0 {
		output.WriteString(fmt.Sprintf("⚠️  WARNINGS: %d\n", len(warnings)))
		output.WriteString(strings.Repeat("-", 50) + "\n")
		for _, v := range warnings {
			output.WriteString(r.formatViolation(v))
		}
		output.WriteString("\n")
	}

	if len(infos) > 0 {
		output.WriteString(fmt.Sprintf("ℹ️  INFO: %d\n", len(infos)))
		output.WriteString(strings.Repeat("-", 50) + "\n")
		for _, v := range infos {
			output.WriteString(r.formatViolation(v))
		}
		output.WriteString("\n")
	}

	// Show filtered violations (exceptions)
	if len(r.filteredViolations) > 0 {
		output.WriteString(fmt.Sprintf("✓ EXCEPTED: %d\n", len(r.filteredViolations)))
		output.WriteString(strings.Repeat("-", 50) + "\n")
		for _, fv := range r.filteredViolations {
			output.WriteString(r.formatFilteredViolation(fv))
		}
		output.WriteString("\n")
	}

	output.WriteString(strings.Repeat("=", 50) + "\n")
	output.WriteString(fmt.Sprintf("Total: %d violations", len(r.violations)))
	if len(r.filteredViolations) > 0 {
		output.WriteString(fmt.Sprintf(" (%d excepted)\n", len(r.filteredViolations)))
	} else {
		output.WriteString("\n")
	}

	return output.String()
}

func (r *Reporter) formatViolation(v config.Violation) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("\n%s:%d:%d\n", v.File, v.Line, v.Column))
	output.WriteString(fmt.Sprintf("  Rule: %s (%s)\n", v.RuleName, v.RuleID))
	output.WriteString(fmt.Sprintf("  Resource: %s.%s\n", v.ResourceType, v.ResourceName))
	output.WriteString(fmt.Sprintf("  Message: %s\n", v.Message))

	if v.Remediation != "" {
		output.WriteString(fmt.Sprintf("  Remediation:\n%s\n", indent(v.Remediation, 4)))
	}

	return output.String()
}

func (r *Reporter) formatFilteredViolation(fv config.FilteredViolation) string {
	var output strings.Builder

	v := fv.Violation
	e := fv.Exception

	output.WriteString(fmt.Sprintf("\n%s:%d:%d\n", v.File, v.Line, v.Column))
	output.WriteString(fmt.Sprintf("  Rule: %s (%s)\n", v.RuleName, v.RuleID))
	output.WriteString(fmt.Sprintf("  Resource: %s.%s\n", v.ResourceType, v.ResourceName))
	output.WriteString(fmt.Sprintf("  Exception Reason: %s\n", e.Reason))
	output.WriteString(fmt.Sprintf("  Approved By: %s\n", e.ApprovedBy))

	if e.Ticket != nil {
		output.WriteString(fmt.Sprintf("  Ticket: %s\n", *e.Ticket))
	}

	if e.ExpiresAt != nil {
		output.WriteString(fmt.Sprintf("  Expires: %s\n", *e.ExpiresAt))
	}

	return output.String()
}

// FormatJSON formats violations as JSON
func (r *Reporter) FormatJSON() (string, error) {
	data, err := json.MarshalIndent(r.violations, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FormatSARIF formats violations as SARIF (Static Analysis Results Interchange Format)
func (r *Reporter) FormatSARIF() (string, error) {
	sarif := map[string]interface{}{
		"version": "2.1.0",
		"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		"runs": []map[string]interface{}{
			{
				"tool": map[string]interface{}{
					"driver": map[string]interface{}{
						"name":           "Terraform Guardian",
						"informationUri": "https://github.com/jonathanhle/planguard",
						"version":        "1.0.0",
						"rules":          r.buildSARIFRules(),
					},
				},
				"results": r.buildSARIFResults(),
			},
		},
	}

	data, err := json.MarshalIndent(sarif, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r *Reporter) buildSARIFRules() []map[string]interface{} {
	// Build unique rules list
	ruleMap := make(map[string]config.Violation)
	for _, v := range r.violations {
		if _, exists := ruleMap[v.RuleID]; !exists {
			ruleMap[v.RuleID] = v
		}
	}

	var rules []map[string]interface{}
	for id, v := range ruleMap {
		rule := map[string]interface{}{
			"id":   id,
			"name": v.RuleName,
			"shortDescription": map[string]interface{}{
				"text": v.RuleName,
			},
			"fullDescription": map[string]interface{}{
				"text": v.Message,
			},
			"defaultConfiguration": map[string]interface{}{
				"level": r.severityToLevel(v.Severity),
			},
		}
		rules = append(rules, rule)
	}

	return rules
}

func (r *Reporter) buildSARIFResults() []map[string]interface{} {
	var results []map[string]interface{}

	for _, v := range r.violations {
		result := map[string]interface{}{
			"ruleId": v.RuleID,
			"level":  r.severityToLevel(v.Severity),
			"message": map[string]interface{}{
				"text": v.Message,
			},
			"locations": []map[string]interface{}{
				{
					"physicalLocation": map[string]interface{}{
						"artifactLocation": map[string]interface{}{
							"uri": v.File,
						},
						"region": map[string]interface{}{
							"startLine":   v.Line,
							"startColumn": v.Column,
						},
					},
				},
			},
		}
		results = append(results, result)
	}

	return results
}

func (r *Reporter) severityToLevel(severity string) string {
	switch severity {
	case "error":
		return "error"
	case "warning":
		return "warning"
	case "info":
		return "note"
	default:
		return "warning"
	}
}

func (r *Reporter) filterBySeverity(severity string) []config.Violation {
	var filtered []config.Violation
	for _, v := range r.violations {
		if v.Severity == severity {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

// ShouldFail determines if the scan should fail based on severity
func (r *Reporter) ShouldFail(failOn string) bool {
	if len(r.violations) == 0 {
		return false
	}

	switch failOn {
	case "error":
		return len(r.filterBySeverity("error")) > 0
	case "warning":
		return len(r.filterBySeverity("warning")) > 0 || len(r.filterBySeverity("error")) > 0
	case "info":
		return len(r.violations) > 0
	default:
		return len(r.filterBySeverity("error")) > 0
	}
}

func indent(text string, spaces int) string {
	prefix := strings.Repeat(" ", spaces)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}
