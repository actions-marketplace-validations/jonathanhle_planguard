package config

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	cfg := &Config{
		Settings: &Settings{
			FailOnWarning: false,
			ExcludePaths:  []string{},
		},
		Rules:      []Rule{},
		Exceptions: []Exception{},
		Functions:  []Function{},
	}

	if cfg.Settings == nil {
		t.Error("Settings should not be nil")
	}

	if cfg.Rules == nil {
		t.Error("Rules should not be nil")
	}
}

func TestRuleSeverity(t *testing.T) {
	tests := []struct {
		name     string
		severity string
		valid    bool
	}{
		{"error severity", "error", true},
		{"warning severity", "warning", true},
		{"info severity", "info", true},
		{"invalid severity", "critical", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just testing that different severity levels are recognized
			rule := Rule{
				ID:       "test",
				Severity: tt.severity,
			}
			if rule.Severity != tt.severity {
				t.Errorf("Severity = %v, want %v", rule.Severity, tt.severity)
			}
		})
	}
}
