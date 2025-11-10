package parser

import (
	"path/filepath"
	"regexp"

	"github.com/jonathanhle/planguard/pkg/config"
)

// ScanContext holds all parsed resources and metadata for scanning
type ScanContext struct {
	// All resources indexed by type
	ResourcesByType map[string][]*config.Resource

	// All resources indexed by file
	ResourcesByFile map[string][]*config.Resource

	// All resources as flat list
	AllResources []*config.Resource

	// Current resource being evaluated (set during rule evaluation)
	CurrentResource *config.Resource

	// Metadata (for GitHub context, etc.)
	Metadata map[string]interface{}
}

// NewScanContext creates a new scan context from resources
func NewScanContext(resources []*config.Resource) *ScanContext {
	ctx := &ScanContext{
		ResourcesByType: make(map[string][]*config.Resource),
		ResourcesByFile: make(map[string][]*config.Resource),
		AllResources:    resources,
		Metadata:        make(map[string]interface{}),
	}

	// Index resources by type
	for _, resource := range resources {
		ctx.ResourcesByType[resource.Type] = append(ctx.ResourcesByType[resource.Type], resource)
		ctx.ResourcesByFile[resource.File] = append(ctx.ResourcesByFile[resource.File], resource)
	}

	return ctx
}

// GetResourcesByType returns all resources matching a type pattern
func (ctx *ScanContext) GetResourcesByType(typePattern string) []*config.Resource {
	var matched []*config.Resource

	// Check for wildcard
	if typePattern == "*" {
		return ctx.AllResources
	}

	// Check for pattern matching (e.g., "aws_*")
	if regexp.MustCompile(`[*?]`).MatchString(typePattern) {
		pattern := "^" + regexp.QuoteMeta(typePattern)
		pattern = regexp.MustCompile(`\\\*`).ReplaceAllString(pattern, ".*")
		pattern = regexp.MustCompile(`\\\?`).ReplaceAllString(pattern, ".")
		re := regexp.MustCompile(pattern + "$")

		for resourceType, resources := range ctx.ResourcesByType {
			if re.MatchString(resourceType) {
				matched = append(matched, resources...)
			}
		}
		return matched
	}

	// Exact match
	return ctx.ResourcesByType[typePattern]
}

// GetResourcesInFile returns all resources in a specific file
func (ctx *ScanContext) GetResourcesInFile(filePath string) []*config.Resource {
	return ctx.ResourcesByFile[filePath]
}

// MatchesPath checks if a file path matches a pattern
func MatchesPath(pattern, path string) bool {
	// Handle ** for recursive matching
	if filepath.IsAbs(pattern) {
		matched, _ := filepath.Match(pattern, path)
		return matched
	}

	// Try matching with glob pattern
	matched, _ := filepath.Match(pattern, filepath.Base(path))
	if matched {
		return true
	}

	// Try matching full path
	matched, _ = filepath.Match(pattern, path)
	return matched
}
