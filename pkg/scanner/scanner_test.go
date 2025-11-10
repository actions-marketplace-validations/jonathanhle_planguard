package scanner

import (
	"strings"
	"testing"
	"time"

	"github.com/planguard/planguard/pkg/config"
	"github.com/planguard/planguard/pkg/parser"
	"github.com/zclconf/go-cty/cty"
)

func TestNewScanner(t *testing.T) {
	cfg := &config.Config{}
	rules := []config.Rule{}
	ctx := parser.NewScanContext([]*config.Resource{})

	scanner := NewScanner(cfg, rules, ctx)
	if scanner == nil {
		t.Fatal("NewScanner returned nil")
	}

	if scanner.config != cfg {
		t.Error("Config not set correctly")
	}

	if scanner.functions == nil {
		t.Error("Functions map should be initialized")
	}
}

func TestScanNoViolations(t *testing.T) {
	resources := []*config.Resource{
		{
			Type: "aws_instance",
			Name: "test",
			Attributes: map[string]cty.Value{
				"instance_type": cty.StringVal("t3.micro"),
			},
		},
	}

	rule := config.Rule{
		ID:           "test",
		Name:         "Test",
		Severity:     "error",
		ResourceType: "aws_instance",
		Conditions: []config.Condition{
			{Expression: "false"}, // Never triggers
		},
		Message: "Test message",
	}

	cfg := &config.Config{}
	ctx := parser.NewScanContext(resources)
	scanner := NewScanner(cfg, []config.Rule{rule}, ctx)

	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(result.Violations) != 0 {
		t.Errorf("Expected 0 violations, got %d", len(result.Violations))
	}
}

func TestScanWithViolations(t *testing.T) {
	resources := []*config.Resource{
		{
			Type: "aws_instance",
			Name: "test",
			Attributes: map[string]cty.Value{
				"instance_type": cty.StringVal("t3.micro"),
			},
			File:   "test.tf",
			Line:   10,
			Column: 5,
		},
	}

	rule := config.Rule{
		ID:           "test_rule",
		Name:         "Test Rule",
		Severity:     "error",
		ResourceType: "aws_instance",
		Conditions: []config.Condition{
			{Expression: "true"}, // Always triggers
		},
		Message: "Test violation",
	}

	cfg := &config.Config{}
	ctx := parser.NewScanContext(resources)
	scanner := NewScanner(cfg, []config.Rule{rule}, ctx)

	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(result.Violations) != 1 {
		t.Fatalf("Expected 1 violation, got %d", len(result.Violations))
	}

	v := result.Violations[0]
	if v.RuleID != "test_rule" {
		t.Errorf("RuleID = %s, want test_rule", v.RuleID)
	}
	if v.File != "test.tf" {
		t.Errorf("File = %s, want test.tf", v.File)
	}
	if v.Line != 10 {
		t.Errorf("Line = %d, want 10", v.Line)
	}
}

func TestScanWithWhenCondition(t *testing.T) {
	resources := []*config.Resource{
		{
			Type: "aws_instance",
			Name: "skip",
			Attributes: map[string]cty.Value{
				"instance_type": cty.StringVal("t3.micro"),
			},
		},
		{
			Type: "aws_instance",
			Name: "check",
			Attributes: map[string]cty.Value{
				"instance_type": cty.StringVal("t3.large"),
			},
		},
	}

	rule := config.Rule{
		ID:           "large_instance",
		Name:         "Large Instance",
		Severity:     "warning",
		ResourceType: "aws_instance",
		When: &config.WhenBlock{
			Expression: `self.instance_type == "t3.large"`,
		},
		Conditions: []config.Condition{
			{Expression: "true"},
		},
		Message: "Large instance detected",
	}

	cfg := &config.Config{}
	ctx := parser.NewScanContext(resources)
	scanner := NewScanner(cfg, []config.Rule{rule}, ctx)

	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	// Should only trigger for the t3.large instance
	if len(result.Violations) != 1 {
		t.Fatalf("Expected 1 violation, got %d", len(result.Violations))
	}

	if result.Violations[0].ResourceName != "check" {
		t.Errorf("Expected violation on 'check' resource, got %s", result.Violations[0].ResourceName)
	}
}

func TestScanWithResourceTypeFilter(t *testing.T) {
	resources := []*config.Resource{
		{
			Type: "aws_instance",
			Name: "instance",
		},
		{
			Type: "aws_s3_bucket",
			Name: "bucket",
		},
	}

	rule := config.Rule{
		ID:           "s3_only",
		Name:         "S3 Only",
		Severity:     "error",
		ResourceType: "aws_s3_bucket",
		Conditions: []config.Condition{
			{Expression: "true"},
		},
		Message: "S3 rule",
	}

	cfg := &config.Config{}
	ctx := parser.NewScanContext(resources)
	scanner := NewScanner(cfg, []config.Rule{rule}, ctx)

	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(result.Violations) != 1 {
		t.Fatalf("Expected 1 violation, got %d", len(result.Violations))
	}

	if result.Violations[0].ResourceType != "aws_s3_bucket" {
		t.Errorf("Expected aws_s3_bucket violation, got %s", result.Violations[0].ResourceType)
	}
}

func TestScanWithRemediation(t *testing.T) {
	resources := []*config.Resource{
		{
			Type: "aws_instance",
			Name: "test",
		},
	}

	remediation := "Fix by setting instance_type = \"t3.micro\""
	rule := config.Rule{
		ID:           "test",
		Name:         "Test",
		Severity:     "error",
		ResourceType: "aws_instance",
		Conditions: []config.Condition{
			{Expression: "true"},
		},
		Message:     "Test",
		Remediation: &remediation,
	}

	cfg := &config.Config{}
	ctx := parser.NewScanContext(resources)
	scanner := NewScanner(cfg, []config.Rule{rule}, ctx)

	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(result.Violations) != 1 {
		t.Fatalf("Expected 1 violation, got %d", len(result.Violations))
	}

	if result.Violations[0].Remediation != remediation {
		t.Errorf("Remediation not set correctly")
	}
}

func TestFilterExceptions(t *testing.T) {
	violations := []config.Violation{
		{
			RuleID:       "excepted_rule",
			ResourceName: "test",
			File:         "test.tf",
		},
		{
			RuleID:       "not_excepted",
			ResourceName: "other",
			File:         "other.tf",
		},
	}

	cfg := &config.Config{
		Exceptions: []config.Exception{
			{
				Rules:      []string{"excepted_rule"},
				Reason:     "Legacy system",
				ApprovedBy: "admin@example.com",
			},
		},
	}

	ctx := parser.NewScanContext([]*config.Resource{})
	scanner := NewScanner(cfg, []config.Rule{}, ctx)

	filtered, excepted := scanner.filterExceptions(violations)

	if len(filtered) != 1 {
		t.Errorf("Expected 1 filtered violation, got %d", len(filtered))
	}

	if len(excepted) != 1 {
		t.Errorf("Expected 1 excepted violation, got %d", len(excepted))
	}

	if filtered[0].RuleID != "not_excepted" {
		t.Errorf("Wrong violation filtered")
	}

	if excepted[0].Violation.RuleID != "excepted_rule" {
		t.Errorf("Wrong violation excepted")
	}
}

func TestFilterExceptionsByPath(t *testing.T) {
	violations := []config.Violation{
		{
			RuleID:       "test",
			ResourceName: "r1",
			File:         "legacy/test.tf",
		},
		{
			RuleID:       "test",
			ResourceName: "r2",
			File:         "new/test.tf",
		},
	}

	cfg := &config.Config{
		Exceptions: []config.Exception{
			{
				Rules:      []string{"test"},
				Paths:      []string{"legacy/*.tf"},
				Reason:     "Legacy code",
				ApprovedBy: "admin@example.com",
			},
		},
	}

	ctx := parser.NewScanContext([]*config.Resource{})
	scanner := NewScanner(cfg, []config.Rule{}, ctx)

	filtered, excepted := scanner.filterExceptions(violations)

	if len(filtered) != 1 {
		t.Errorf("Expected 1 filtered violation, got %d", len(filtered))
	}

	if len(excepted) != 1 {
		t.Errorf("Expected 1 excepted violation, got %d", len(excepted))
	}

	if filtered[0].File != "new/test.tf" {
		t.Errorf("Wrong file filtered")
	}
}

func TestFilterExceptionsByResourceName(t *testing.T) {
	violations := []config.Violation{
		{
			RuleID:       "test",
			ResourceName: "legacy_resource",
			File:         "test.tf",
		},
		{
			RuleID:       "test",
			ResourceName: "new_resource",
			File:         "test.tf",
		},
	}

	cfg := &config.Config{
		Exceptions: []config.Exception{
			{
				Rules:         []string{"test"},
				ResourceNames: []string{"legacy_*"},
				Reason:        "Legacy resource",
				ApprovedBy:    "admin@example.com",
			},
		},
	}

	ctx := parser.NewScanContext([]*config.Resource{})
	scanner := NewScanner(cfg, []config.Rule{}, ctx)

	filtered, excepted := scanner.filterExceptions(violations)

	if len(filtered) != 1 {
		t.Errorf("Expected 1 filtered violation, got %d", len(filtered))
	}

	if len(excepted) != 1 {
		t.Errorf("Expected 1 excepted violation, got %d", len(excepted))
	}

	if filtered[0].ResourceName != "new_resource" {
		t.Errorf("Wrong resource filtered")
	}
}

func TestFilterExceptionsExpired(t *testing.T) {
	violations := []config.Violation{
		{
			RuleID:       "test",
			ResourceName: "resource",
			File:         "test.tf",
		},
	}

	// Exception expired yesterday
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	cfg := &config.Config{
		Exceptions: []config.Exception{
			{
				Rules:      []string{"test"},
				Reason:     "Temporary",
				ApprovedBy: "admin@example.com",
				ExpiresAt:  &yesterday,
			},
		},
	}

	ctx := parser.NewScanContext([]*config.Resource{})
	scanner := NewScanner(cfg, []config.Rule{}, ctx)

	filtered, excepted := scanner.filterExceptions(violations)

	// Expired exception should not be applied
	if len(filtered) != 1 {
		t.Errorf("Expected 1 filtered violation (expired), got %d", len(filtered))
	}

	if len(excepted) != 0 {
		t.Errorf("Expected 0 excepted violations (expired), got %d", len(excepted))
	}
}

func TestFilterExceptionsNotExpired(t *testing.T) {
	violations := []config.Violation{
		{
			RuleID:       "test",
			ResourceName: "resource",
			File:         "test.tf",
		},
	}

	// Exception expires tomorrow
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	cfg := &config.Config{
		Exceptions: []config.Exception{
			{
				Rules:      []string{"test"},
				Reason:     "Temporary",
				ApprovedBy: "admin@example.com",
				ExpiresAt:  &tomorrow,
			},
		},
	}

	ctx := parser.NewScanContext([]*config.Resource{})
	scanner := NewScanner(cfg, []config.Rule{}, ctx)

	filtered, excepted := scanner.filterExceptions(violations)

	// Not expired - should be applied
	if len(filtered) != 0 {
		t.Errorf("Expected 0 filtered violations, got %d", len(filtered))
	}

	if len(excepted) != 1 {
		t.Errorf("Expected 1 excepted violation, got %d", len(excepted))
	}
}

func TestEvaluateExpressionSimple(t *testing.T) {
	resource := &config.Resource{
		Type: "aws_instance",
		Name: "test",
		Attributes: map[string]cty.Value{
			"instance_type": cty.StringVal("t3.micro"),
		},
	}

	ctx := parser.NewScanContext([]*config.Resource{resource})
	scanner := NewScanner(&config.Config{}, []config.Rule{}, ctx)

	result, err := scanner.evaluateExpression("true", resource)
	if err != nil {
		t.Fatalf("evaluateExpression() error = %v", err)
	}

	if !result {
		t.Error("Expected true result")
	}
}

func TestEvaluateExpressionWithSelf(t *testing.T) {
	resource := &config.Resource{
		Type: "aws_instance",
		Name: "test",
		Attributes: map[string]cty.Value{
			"instance_type": cty.StringVal("t3.large"),
		},
	}

	ctx := parser.NewScanContext([]*config.Resource{resource})
	scanner := NewScanner(&config.Config{}, []config.Rule{}, ctx)

	result, err := scanner.evaluateExpression(`self.instance_type == "t3.large"`, resource)
	if err != nil {
		t.Fatalf("evaluateExpression() error = %v", err)
	}

	if !result {
		t.Error("Expected true result for matching instance_type")
	}
}

func TestEvaluateExpressionInvalid(t *testing.T) {
	resource := &config.Resource{
		Type: "aws_instance",
		Name: "test",
	}

	ctx := parser.NewScanContext([]*config.Resource{resource})
	scanner := NewScanner(&config.Config{}, []config.Rule{}, ctx)

	_, err := scanner.evaluateExpression("invalid {{{ syntax", resource)
	if err == nil {
		t.Error("Expected error for invalid expression")
	}
}

func TestEvaluateExpressionNonBoolean(t *testing.T) {
	resource := &config.Resource{
		Type: "aws_instance",
		Name: "test",
	}

	ctx := parser.NewScanContext([]*config.Resource{resource})
	scanner := NewScanner(&config.Config{}, []config.Rule{}, ctx)

	_, err := scanner.evaluateExpression(`"string"`, resource)
	if err == nil {
		t.Error("Expected error for non-boolean expression")
	}

	if !strings.Contains(err.Error(), "must return boolean") {
		t.Errorf("Expected 'must return boolean' error, got: %v", err)
	}
}

func TestResourceToCtyValue(t *testing.T) {
	resource := &config.Resource{
		Type:   "aws_instance",
		Name:   "test",
		File:   "test.tf",
		Line:   10,
		Column: 5,
		Attributes: map[string]cty.Value{
			"instance_type": cty.StringVal("t3.micro"),
			"ami":           cty.StringVal("ami-12345"),
		},
	}

	value := resourceToCtyValue(resource)

	if value.Type().IsObjectType() == false {
		t.Error("Expected object type")
	}

	// Check metadata fields
	typeVal := value.GetAttr("type")
	if typeVal.AsString() != "aws_instance" {
		t.Error("Type not set correctly")
	}

	nameVal := value.GetAttr("name")
	if nameVal.AsString() != "test" {
		t.Error("Name not set correctly")
	}

	fileVal := value.GetAttr("file")
	if fileVal.AsString() != "test.tf" {
		t.Error("File not set correctly")
	}

	// Check attributes
	instanceTypeVal := value.GetAttr("instance_type")
	if instanceTypeVal.AsString() != "t3.micro" {
		t.Error("instance_type not set correctly")
	}
}

func TestScanMultipleRules(t *testing.T) {
	resources := []*config.Resource{
		{
			Type: "aws_instance",
			Name: "test",
		},
	}

	rules := []config.Rule{
		{
			ID:           "rule1",
			Name:         "Rule 1",
			Severity:     "error",
			ResourceType: "aws_instance",
			Conditions:   []config.Condition{{Expression: "true"}},
			Message:      "Rule 1 violation",
		},
		{
			ID:           "rule2",
			Name:         "Rule 2",
			Severity:     "warning",
			ResourceType: "aws_instance",
			Conditions:   []config.Condition{{Expression: "true"}},
			Message:      "Rule 2 violation",
		},
	}

	cfg := &config.Config{}
	ctx := parser.NewScanContext(resources)
	scanner := NewScanner(cfg, rules, ctx)

	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	// Should have 2 violations (one from each rule)
	if len(result.Violations) != 2 {
		t.Errorf("Expected 2 violations, got %d", len(result.Violations))
	}
}

func TestScanRuleError(t *testing.T) {
	resources := []*config.Resource{
		{
			Type: "aws_instance",
			Name: "test",
		},
	}

	rule := config.Rule{
		ID:           "invalid",
		Name:         "Invalid",
		Severity:     "error",
		ResourceType: "aws_instance",
		Conditions: []config.Condition{
			{Expression: "invalid {{{ syntax"},
		},
		Message: "Test",
	}

	cfg := &config.Config{}
	ctx := parser.NewScanContext(resources)
	scanner := NewScanner(cfg, []config.Rule{rule}, ctx)

	_, err := scanner.Scan()
	if err == nil {
		t.Error("Expected error for invalid rule expression")
	}

	if !strings.Contains(err.Error(), "error scanning rule") {
		t.Errorf("Expected 'error scanning rule' in error, got: %v", err)
	}
}

func TestScanWhenConditionError(t *testing.T) {
	resources := []*config.Resource{
		{
			Type: "aws_instance",
			Name: "test",
		},
	}

	rule := config.Rule{
		ID:           "invalid",
		Name:         "Invalid",
		Severity:     "error",
		ResourceType: "aws_instance",
		When: &config.WhenBlock{
			Expression: "invalid {{{ syntax",
		},
		Conditions: []config.Condition{
			{Expression: "true"},
		},
		Message: "Test",
	}

	cfg := &config.Config{}
	ctx := parser.NewScanContext(resources)
	scanner := NewScanner(cfg, []config.Rule{rule}, ctx)

	_, err := scanner.Scan()
	if err == nil {
		t.Error("Expected error for invalid when condition")
	}
}

func TestFindExceptionNoMatch(t *testing.T) {
	violation := config.Violation{
		RuleID:       "test",
		ResourceName: "resource",
		File:         "test.tf",
	}

	cfg := &config.Config{
		Exceptions: []config.Exception{
			{
				Rules:      []string{"other_rule"}, // Different rule
				Reason:     "Test",
				ApprovedBy: "admin",
			},
		},
	}

	ctx := parser.NewScanContext([]*config.Resource{})
	scanner := NewScanner(cfg, []config.Rule{}, ctx)

	exception, found := scanner.findException(violation)
	if found {
		t.Error("Should not find exception for different rule")
	}
	if exception != nil {
		t.Error("Exception should be nil when not found")
	}
}

func TestFilterExceptionsInvalidExpirationDate(t *testing.T) {
	violations := []config.Violation{
		{
			RuleID:       "test",
			ResourceName: "resource",
			File:         "test.tf",
		},
	}

	invalidDate := "not-a-date"

	cfg := &config.Config{
		Exceptions: []config.Exception{
			{
				Rules:      []string{"test"},
				Reason:     "Test",
				ApprovedBy: "admin",
				ExpiresAt:  &invalidDate,
			},
		},
	}

	ctx := parser.NewScanContext([]*config.Resource{})
	scanner := NewScanner(cfg, []config.Rule{}, ctx)

	filtered, excepted := scanner.filterExceptions(violations)

	// Invalid date format should be treated as valid (parsing error ignored)
	if len(excepted) != 1 {
		t.Errorf("Expected 1 excepted violation (invalid date ignored), got %d", len(excepted))
	}

	if len(filtered) != 0 {
		t.Errorf("Expected 0 filtered violations, got %d", len(filtered))
	}
}
