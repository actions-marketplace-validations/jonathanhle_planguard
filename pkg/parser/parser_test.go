package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/planguard/planguard/pkg/config"
)

func TestNewParser(t *testing.T) {
	p := NewParser()
	if p == nil {
		t.Fatal("NewParser() returned nil")
	}

	if p.hclParser == nil {
		t.Error("HCL parser should be initialized")
	}
}

func TestParseFile(t *testing.T) {
	// Create temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.tf")

	content := `
resource "aws_instance" "example" {
  ami           = "ami-12345678"
  instance_type = "t3.micro"

  tags = {
    Name = "test-instance"
  }
}
`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	p := NewParser()
	file, err := p.ParseFile(testFile)

	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	if file == nil {
		t.Fatal("ParseFile() returned nil file")
	}
}

func TestParseFileNotFound(t *testing.T) {
	p := NewParser()
	_, err := p.ParseFile("/nonexistent/file.tf")

	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestParseFileInvalidHCL(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.tf")

	content := `
resource "aws_instance" {
  invalid syntax here {{{
}
`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	p := NewParser()
	_, err = p.ParseFile(testFile)

	if err == nil {
		t.Error("Expected error for invalid HCL")
	}
}

func TestParseDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	file1 := filepath.Join(tmpDir, "main.tf")
	file2 := filepath.Join(tmpDir, "variables.tf")
	nonTfFile := filepath.Join(tmpDir, "readme.md")

	tfContent := `
resource "aws_s3_bucket" "example" {
  bucket = "test-bucket"
}
`

	if err := os.WriteFile(file1, []byte(tfContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte(tfContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(nonTfFile, []byte("# README"), 0644); err != nil {
		t.Fatal(err)
	}

	p := NewParser()
	files, err := p.ParseDirectory(tmpDir, []string{})

	if err != nil {
		t.Fatalf("ParseDirectory() error = %v", err)
	}

	// Should parse 2 .tf files, not the .md file
	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}
}

func TestParseDirectoryWithExcludes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create subdirectory
	terraformDir := filepath.Join(tmpDir, ".terraform")
	if err := os.MkdirAll(terraformDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create files
	mainFile := filepath.Join(tmpDir, "main.tf")
	excludedFile := filepath.Join(terraformDir, "excluded.tf")

	tfContent := `resource "aws_instance" "test" {}`

	if err := os.WriteFile(mainFile, []byte(tfContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(excludedFile, []byte(tfContent), 0644); err != nil {
		t.Fatal(err)
	}

	p := NewParser()
	files, err := p.ParseDirectory(tmpDir, []string{".terraform"})

	if err != nil {
		t.Fatalf("ParseDirectory() error = %v", err)
	}

	// Should only parse main.tf, not the file in .terraform
	if len(files) != 1 {
		t.Errorf("Expected 1 file (excluding .terraform), got %d", len(files))
	}
}

func TestParseDirectoryEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	p := NewParser()
	files, err := p.ParseDirectory(tmpDir, []string{})

	if err != nil {
		t.Fatalf("ParseDirectory() error = %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected 0 files in empty directory, got %d", len(files))
	}
}

func TestExtractResources(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.tf")

	content := `
resource "aws_instance" "web" {
  ami           = "ami-12345678"
  instance_type = "t3.micro"
}

resource "aws_s3_bucket" "data" {
  bucket = "my-bucket"
}

data "aws_ami" "ubuntu" {
  most_recent = true
}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	p := NewParser()
	files, err := p.ParseDirectory(tmpDir, []string{})
	if err != nil {
		t.Fatalf("ParseDirectory() error = %v", err)
	}

	resources, err := ExtractResources(files)
	if err != nil {
		t.Fatalf("ExtractResources() error = %v", err)
	}

	// Should extract 2 resources + 1 data source = 3 total
	if len(resources) != 3 {
		t.Errorf("Expected 3 resources, got %d", len(resources))
	}

	// Verify first resource
	if resources[0].Type != "aws_instance" {
		t.Errorf("First resource type = %s, want aws_instance", resources[0].Type)
	}

	if resources[0].Name != "web" {
		t.Errorf("First resource name = %s, want web", resources[0].Name)
	}

	// Verify attributes were extracted
	if len(resources[0].Attributes) == 0 {
		t.Error("Resource should have attributes")
	}
}

func TestExtractResourcesWithComplexAttributes(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "complex.tf")

	content := `
resource "aws_instance" "complex" {
  ami           = "ami-12345678"
  instance_type = "t3.micro"

  tags = {
    Name        = "complex-instance"
    Environment = "test"
  }

  security_groups = ["sg-12345", "sg-67890"]
}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	p := NewParser()
	files, err := p.ParseDirectory(tmpDir, []string{})
	if err != nil {
		t.Fatalf("ParseDirectory() error = %v", err)
	}

	resources, err := ExtractResources(files)
	if err != nil {
		t.Fatalf("ExtractResources() error = %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("Expected 1 resource, got %d", len(resources))
	}

	resource := resources[0]

	// Verify file location info
	if resource.File != testFile {
		t.Errorf("Resource file = %s, want %s", resource.File, testFile)
	}

	if resource.Line == 0 {
		t.Error("Resource line should be set")
	}

	// Verify raw expressions are captured
	if len(resource.RawExprs) == 0 {
		t.Error("Resource should have raw expressions")
	}
}

func TestExtractResourcesEmpty(t *testing.T) {
	resources, err := ExtractResources(map[string]*hcl.File{})

	if err != nil {
		t.Fatalf("ExtractResources() error = %v", err)
	}

	if len(resources) != 0 {
		t.Errorf("Expected 0 resources from empty files, got %d", len(resources))
	}
}

func TestNewScanContext(t *testing.T) {
	resources := []*config.Resource{
		{Type: "aws_instance", Name: "test1"},
		{Type: "aws_s3_bucket", Name: "test2"},
		{Type: "aws_instance", Name: "test3"},
	}

	ctx := NewScanContext(resources)

	if ctx == nil {
		t.Fatal("NewScanContext() returned nil")
	}

	if len(ctx.AllResources) != 3 {
		t.Errorf("Expected 3 resources, got %d", len(ctx.AllResources))
	}
}

func TestGetResourcesByType(t *testing.T) {
	resources := []*config.Resource{
		{Type: "aws_instance", Name: "test1"},
		{Type: "aws_s3_bucket", Name: "test2"},
		{Type: "aws_instance", Name: "test3"},
	}

	ctx := NewScanContext(resources)

	instances := ctx.GetResourcesByType("aws_instance")
	if len(instances) != 2 {
		t.Errorf("Expected 2 aws_instance resources, got %d", len(instances))
	}

	buckets := ctx.GetResourcesByType("aws_s3_bucket")
	if len(buckets) != 1 {
		t.Errorf("Expected 1 aws_s3_bucket resource, got %d", len(buckets))
	}

	nonexistent := ctx.GetResourcesByType("aws_lambda_function")
	if len(nonexistent) != 0 {
		t.Errorf("Expected 0 nonexistent resources, got %d", len(nonexistent))
	}
}

func TestGetResourcesByTypeWildcard(t *testing.T) {
	resources := []*config.Resource{
		{Type: "aws_instance", Name: "test1"},
		{Type: "aws_s3_bucket", Name: "test2"},
		{Type: "azurerm_virtual_machine", Name: "test3"},
	}

	ctx := NewScanContext(resources)

	// Test wildcard matching
	awsResources := ctx.GetResourcesByType("aws_*")
	if len(awsResources) != 2 {
		t.Errorf("Expected 2 aws_* resources, got %d", len(awsResources))
	}

	allResources := ctx.GetResourcesByType("*")
	if len(allResources) != 3 {
		t.Errorf("Expected 3 resources with *, got %d", len(allResources))
	}
}

func TestMatchesPath(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		path     string
		expected bool
	}{
		{
			name:     "exact match",
			pattern:  "test.tf",
			path:     "test.tf",
			expected: true,
		},
		{
			name:     "wildcard match",
			pattern:  "*.tf",
			path:     "main.tf",
			expected: true,
		},
		{
			name:     "no match",
			pattern:  "*.hcl",
			path:     "main.tf",
			expected: false,
		},
		{
			name:     "path match",
			pattern:  "legacy/*.tf",
			path:     "legacy/old.tf",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchesPath(tt.pattern, tt.path)
			if result != tt.expected {
				t.Errorf("MatchesPath(%q, %q) = %v, want %v", tt.pattern, tt.path, result, tt.expected)
			}
		})
	}
}
