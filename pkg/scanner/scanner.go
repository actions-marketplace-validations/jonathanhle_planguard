package scanner

import (
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/planguard/planguard/pkg/config"
	"github.com/planguard/planguard/pkg/functions"
	"github.com/planguard/planguard/pkg/parser"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// Scanner performs security scanning on Terraform files
type Scanner struct {
	config    *config.Config
	rules     []config.Rule
	context   *parser.ScanContext
	functions map[string]function.Function
}

// NewScanner creates a new scanner instance
func NewScanner(cfg *config.Config, rules []config.Rule, ctx *parser.ScanContext) *Scanner {
	return &Scanner{
		config:    cfg,
		rules:     rules,
		context:   ctx,
		functions: functions.BuildFunctions(ctx),
	}
}

// ScanResult contains both violations and filtered violations
type ScanResult struct {
	Violations         []config.Violation
	FilteredViolations []config.FilteredViolation
}

// Scan performs the security scan
func (s *Scanner) Scan() (*ScanResult, error) {
	var violations []config.Violation

	// Scan each rule
	for _, rule := range s.rules {
		ruleViolations, err := s.scanRule(rule)
		if err != nil {
			return nil, fmt.Errorf("error scanning rule %s: %w", rule.ID, err)
		}
		violations = append(violations, ruleViolations...)
	}

	// Filter exceptions and track filtered violations
	filtered, excepted := s.filterExceptions(violations)

	return &ScanResult{
		Violations:         filtered,
		FilteredViolations: excepted,
	}, nil
}

func (s *Scanner) scanRule(rule config.Rule) ([]config.Violation, error) {
	var violations []config.Violation

	// Get resources matching the resource type
	resources := s.context.GetResourcesByType(rule.ResourceType)

	for _, resource := range resources {
		// Set current resource in context
		s.context.CurrentResource = resource

		// Check when condition
		if rule.When != nil {
			shouldRun, err := s.evaluateExpression(rule.When.Expression, resource)
			if err != nil {
				return nil, fmt.Errorf("error evaluating when condition: %w", err)
			}
			if !shouldRun {
				continue
			}
		}

		// Check all conditions
		violated := false
		for _, condition := range rule.Conditions {
			result, err := s.evaluateExpression(condition.Expression, resource)
			if err != nil {
				return nil, fmt.Errorf("error evaluating condition: %w", err)
			}

			// If condition is true, it's a violation
			if result {
				violated = true
				break
			}
		}

		if violated {
			violation := config.Violation{
				RuleID:       rule.ID,
				RuleName:     rule.Name,
				Severity:     rule.Severity,
				Message:      rule.Message,
				File:         resource.File,
				Line:         resource.Line,
				Column:       resource.Column,
				ResourceType: resource.Type,
				ResourceName: resource.Name,
			}

			if rule.Remediation != nil {
				violation.Remediation = *rule.Remediation
			}

			violations = append(violations, violation)
		}
	}

	return violations, nil
}

func (s *Scanner) evaluateExpression(exprStr string, resource *config.Resource) (bool, error) {
	// Parse the expression
	expr, diags := hclsyntax.ParseExpression([]byte(exprStr), "", hcl.Pos{})
	if diags.HasErrors() {
		return false, fmt.Errorf("invalid expression: %s", diags.Error())
	}

	// Build evaluation context
	evalCtx := &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"self": resourceToCtyValue(resource),
		},
		Functions: s.functions,
	}

	// Evaluate expression
	value, diags := expr.Value(evalCtx)
	if diags.HasErrors() {
		return false, fmt.Errorf("evaluation error: %s", diags.Error())
	}

	// Convert to boolean
	if value.Type() == cty.Bool {
		return value.True(), nil
	}

	return false, fmt.Errorf("expression must return boolean, got %s", value.Type().FriendlyName())
}

func (s *Scanner) filterExceptions(violations []config.Violation) ([]config.Violation, []config.FilteredViolation) {
	var filtered []config.Violation
	var excepted []config.FilteredViolation

	for _, violation := range violations {
		exception, isExcepted := s.findException(violation)
		if isExcepted {
			// Log real-time feedback when exception is applied
			fmt.Fprintf(os.Stderr, "✓ Exception applied: %s.%s - %s (Reason: %s)\n",
				violation.ResourceType,
				violation.ResourceName,
				violation.RuleID,
				exception.Reason)

			excepted = append(excepted, config.FilteredViolation{
				Violation: violation,
				Exception: *exception,
			})
		} else {
			filtered = append(filtered, violation)
		}
	}

	return filtered, excepted
}

func (s *Scanner) findException(violation config.Violation) (*config.Exception, bool) {
	for _, exception := range s.config.Exceptions {
		// Check if rule matches
		ruleMatched := false
		for _, ruleID := range exception.Rules {
			if ruleID == violation.RuleID {
				ruleMatched = true
				break
			}
		}
		if !ruleMatched {
			continue
		}

		// Check if path matches
		if len(exception.Paths) > 0 {
			pathMatched := false
			for _, pattern := range exception.Paths {
				if parser.MatchesPath(pattern, violation.File) {
					pathMatched = true
					break
				}
			}
			if !pathMatched {
				continue
			}
		}

		// Check if resource name matches
		if len(exception.ResourceNames) > 0 {
			nameMatched := false
			for _, pattern := range exception.ResourceNames {
				if parser.MatchesPath(pattern, violation.ResourceName) {
					nameMatched = true
					break
				}
			}
			if !nameMatched {
				continue
			}
		}

		// Check expiration
		if exception.ExpiresAt != nil {
			expiryDate, err := time.Parse("2006-01-02", *exception.ExpiresAt)
			if err == nil && time.Now().After(expiryDate) {
				// Exception expired
				continue
			}
		}

		// All checks passed - exception applies
		return &exception, true
	}

	return nil, false
}

func resourceToCtyValue(resource *config.Resource) cty.Value {
	attrs := make(map[string]cty.Value)

	// Add metadata
	attrs["type"] = cty.StringVal(resource.Type)
	attrs["name"] = cty.StringVal(resource.Name)
	attrs["file"] = cty.StringVal(resource.File)
	attrs["line"] = cty.NumberIntVal(int64(resource.Line))

	// Add all resource attributes
	for key, val := range resource.Attributes {
		attrs[key] = val
	}

	return cty.ObjectVal(attrs)
}
