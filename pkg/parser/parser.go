package parser

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/jonathanhle/planguard/pkg/config"
	"github.com/zclconf/go-cty/cty"
)

// Parser handles parsing of Terraform files
type Parser struct {
	hclParser *hclparse.Parser
}

// NewParser creates a new parser instance
func NewParser() *Parser {
	return &Parser{
		hclParser: hclparse.NewParser(),
	}
}

// ParseFile parses a single Terraform file
func (p *Parser) ParseFile(path string) (*hcl.File, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	file, diags := p.hclParser.ParseHCL(content, path)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse %s: %s", path, diags.Error())
	}

	return file, nil
}

// ParseDirectory recursively parses all .tf files in a directory
func (p *Parser) ParseDirectory(dir string, excludePatterns []string) (map[string]*hcl.File, error) {
	files := make(map[string]*hcl.File)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Check if directory should be excluded
			for _, pattern := range excludePatterns {
				matched, _ := filepath.Match(pattern, filepath.Base(path))
				if matched {
					return filepath.SkipDir
				}
			}
			return nil
		}

		if filepath.Ext(path) != ".tf" {
			return nil
		}

		// Check if file should be excluded
		for _, pattern := range excludePatterns {
			matched, _ := filepath.Match(pattern, path)
			if matched {
				return nil
			}
		}

		file, err := p.ParseFile(path)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		files[path] = file
		return nil
	})

	return files, err
}

// ExtractResources extracts all resources from parsed HCL files
func ExtractResources(files map[string]*hcl.File) ([]*config.Resource, error) {
	var resources []*config.Resource

	for path, file := range files {
		fileResources, err := extractResourcesFromFile(file, path)
		if err != nil {
			return nil, err
		}
		resources = append(resources, fileResources...)
	}

	return resources, nil
}

func extractResourcesFromFile(file *hcl.File, path string) ([]*config.Resource, error) {
	var resources []*config.Resource

	content, _, diags := file.Body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "resource",
				LabelNames: []string{"type", "name"},
			},
			{
				Type:       "data",
				LabelNames: []string{"type", "name"},
			},
		},
	})

	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse content: %s", diags.Error())
	}

	for _, block := range content.Blocks {
		if block.Type != "resource" && block.Type != "data" {
			continue
		}

		resource := &config.Resource{
			Type:       block.Labels[0],
			Name:       block.Labels[1],
			File:       path,
			Line:       block.DefRange.Start.Line,
			Column:     block.DefRange.Start.Column,
			Labels:     block.Labels,
			Attributes: make(map[string]cty.Value),
			RawExprs:   make(map[string]hcl.Expression),
		}

		// Extract attributes
		attrs, diags := block.Body.JustAttributes()
		if !diags.HasErrors() {
			for name, attr := range attrs {
				// Store raw expression for function call detection
				resource.RawExprs[name] = attr.Expr

				// Also evaluate and store the value
				val, diags := attr.Expr.Value(nil)
				if !diags.HasErrors() {
					resource.Attributes[name] = val
				}
			}
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
