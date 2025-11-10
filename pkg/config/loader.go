package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

// LoadConfig loads the guardian configuration from a file
func LoadConfig(configPath string) (*Config, error) {
	var config Config

	err := hclsimple.DecodeFile(configPath, nil, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Set defaults
	if config.Settings == nil {
		config.Settings = &Settings{
			FailOnWarning: false,
			ExcludePaths:  []string{},
		}
	}

	return &config, nil
}

// LoadRules loads rules from one or more HCL files
func LoadRules(rulesPaths []string) ([]Rule, error) {
	var allRules []Rule

	for _, path := range rulesPaths {
		// Check if path is a pattern
		matches, err := filepath.Glob(path)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %s: %w", path, err)
		}

		if len(matches) == 0 {
			// Try as literal path
			matches = []string{path}
		}

		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil {
				continue
			}

			if info.IsDir() {
				// Load all .hcl files in directory
				files, err := filepath.Glob(filepath.Join(match, "*.hcl"))
				if err != nil {
					continue
				}
				matches = append(matches, files...)
				continue
			}

			// Load rules from file
			var fileConfig struct {
				Rules []Rule `hcl:"rule,block"`
			}

			err = hclsimple.DecodeFile(match, nil, &fileConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to load rules from %s: %w", match, err)
			}

			allRules = append(allRules, fileConfig.Rules...)
		}
	}

	return allRules, nil
}

// LoadDefaultRules loads built-in default rules
func LoadDefaultRules(rulesDir string) ([]Rule, error) {
	if rulesDir == "" {
		// Use embedded rules or skip
		return []Rule{}, nil
	}

	var patterns []string

	// Load rules from root directory
	rootPattern := filepath.Join(rulesDir, "*.hcl")
	patterns = append(patterns, rootPattern)

	// Load rules from provider subdirectories
	providers := []string{"aws", "azure", "common"}
	for _, provider := range providers {
		pattern := filepath.Join(rulesDir, provider, "*.hcl")
		patterns = append(patterns, pattern)
	}

	return LoadRules(patterns)
}
