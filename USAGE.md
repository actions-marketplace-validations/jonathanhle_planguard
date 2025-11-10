# Planguard - Usage Guide

## Installation

### Option 1: From Source

```bash
# Clone the repository
git clone https://github.com/your-org/planguard.git
cd planguard

# Build
make build

# The binary will be in bin/planguard
./bin/planguard -version
```

### Option 2: Using Go Install

```bash
go install github.com/planguard/planguard/cmd/planguard@latest
```

### Option 3: Download Binary

Download from [Releases](https://github.com/planguard/planguard/releases) page.

## Quick Start

### 1. Initialize Configuration

Create `.planguard/config.hcl` in your Terraform repository:

```bash
mkdir .planguard
```

### 2. Use Default Rules

Planguard ships with 20+ default rules. To use them:

```bash
planguard -directory ./terraform -rules-dir /path/to/planguard/rules
```

Or in your config:

```hcl
# .planguard/config.hcl

settings {
  fail_on_warning = false
}

# Default rules will be automatically loaded
```

### 3. Run Your First Scan

```bash
planguard -config .planguard/config.hcl -directory ./terraform
```

## Examples

### Example 1: Scan with Default Rules

```bash
cd planguard
make run-example
```

This will:
1. Build planguard
2. Run it on the example Terraform files
3. Show violations in text format

Expected output:
```
Found 4 resources in 1 files
🔒 Planguard Scan Results
==================================================

❌ ERRORS: 4
--------------------------------------------------

examples/terraform/main.tf:5:3
  Rule: Prevent public-read S3 buckets (aws_s3_public_read)
  Resource: aws_s3_bucket.public_bucket
  Message: S3 buckets must not have public-read ACL

...
```

### Example 2: JSON Output

```bash
make run-example-json
```

Returns structured JSON for CI/CD integration.

### Example 3: SARIF for GitHub

```bash
make run-example-sarif > results.sarif
```

Upload to GitHub Security tab.

## Writing Expressions

Planguard expressions use Terraform expression syntax. Choose the right format based on complexity:

### Simple Single-Line Expressions

For basic comparisons without string literals:

```hcl
condition {
  expression = "self.enabled == true"
}
```

### Expressions with Strings (Requires Escaping)

When using string literals in single-line expressions, escape inner quotes:

```hcl
condition {
  expression = "contains_function_call(\"nonsensitive\")"
}

condition {
  expression = "lookup(self.tags, \"Environment\", \"\") == \"prod\""
}
```

### Heredoc Syntax (Recommended for Complex Expressions)

**Use heredoc to eliminate quote escaping** - especially for multi-line expressions:

```hcl
condition {
  expression = <<-EXPR
    contains_function_call("nonsensitive") &&
    lookup(self.tags, "Environment", "") == "prod"
  EXPR
}
```

**Why use heredoc?**
- ✅ No quote escaping needed
- ✅ Multi-line for readability
- ✅ Cleaner for complex logic

### Quick Reference

| Expression Type | Syntax |
|----------------|--------|
| Simple (no strings) | `"self.enabled == true"` |
| With strings (few) | `"lookup(self.tags, \"Env\", \"\")"` |
| Complex or multi-line | `<<-EXPR ... EXPR` |

## Writing Your First Rule

### Step 1: Create a Rule File

Create `.planguard/rules/my-rules.hcl`:

```hcl
rule "require_kms_encryption" {
  name     = "S3 buckets must use KMS encryption"
  severity = "error"

  resource_type = "aws_s3_bucket"

  condition {
    expression = <<-EXPR
      !has(self, 'server_side_encryption_configuration') ||
      !anytrue([
        for rule in self.server_side_encryption_configuration.rule :
        has(rule.apply_server_side_encryption_by_default, 'kms_master_key_id')
      ])
    EXPR
  }

  message = "S3 buckets must use KMS encryption, not AES256"
}
```

### Step 2: Reference in Config

Update `.planguard/config.hcl`:

```hcl
settings {
  fail_on_warning = false
}

# Import your custom rules
# (Rules in config.hcl are automatically loaded)

rule "require_kms_encryption" {
  # ... rule definition from above
}
```

### Step 3: Test Your Rule

Create a test Terraform file that should trigger the violation:

```hcl
# test.tf
resource "aws_s3_bucket" "test" {
  bucket = "my-bucket"
  # No KMS encryption - should fail
}
```

Run planguard:

```bash
planguard -config .planguard/config.hcl -directory .
```

## Common Patterns

### Pattern 1: Environment-Specific Rules

```hcl
rule "prod_multi_az" {
  name     = "Production RDS must be multi-AZ"
  severity = "error"
  
  resource_type = "aws_db_instance"
  
  when {
    expression = "lookup(self.tags, 'Environment', '') == 'prod'"
  }
  
  condition {
    expression = "!has(self, 'multi_az') || self.multi_az != true"
  }
  
  message = "Production RDS instances must be multi-AZ for high availability"
}
```

### Pattern 2: Cross-Resource Validation

```hcl
rule "alb_has_target_group" {
  name     = "ALBs must have target groups"
  severity = "error"
  
  resource_type = "aws_lb"
  
  condition {
    expression = <<-EXPR
      length([
        for tg in resources("aws_lb_target_group") :
        tg if contains(
          [for listener in resources("aws_lb_listener") :
           listener.load_balancer_arn],
          self.arn
        )
      ]) == 0
    EXPR
  }
  
  message = "Load balancer has no associated target groups"
}
```

### Pattern 3: JSON Policy Analysis

```hcl
rule "s3_policy_no_public_access" {
  name     = "S3 bucket policies must not allow public access"
  severity = "error"
  
  resource_type = "aws_s3_bucket_policy"
  
  condition {
    expression = <<-EXPR
      anytrue([
        for stmt in jsondecode(self.policy).Statement :
        contains(try(tolist(stmt.Principal), [stmt.Principal]), "*") ||
        has(stmt, 'Principal') && stmt.Principal == "*"
      ])
    EXPR
  }
  
  message = "S3 bucket policy allows public access"
}
```

## Testing Rules

### Manual Testing

1. Create test Terraform files that should trigger violations
2. Run planguard on them
3. Verify violations are detected

```bash
# Good test (should pass)
planguard -config test-config.hcl -directory good-examples/

# Bad test (should fail)
planguard -config test-config.hcl -directory bad-examples/
```

### Automated Testing

Create a test script:

```bash
#!/bin/bash
# test-rules.sh

echo "Testing rule: no_public_s3"
planguard -config .planguard/config.hcl -directory test/s3-public/ | grep "aws_s3_public_read"
if [ $? -eq 0 ]; then
  echo "✅ Rule detected violation correctly"
else
  echo "❌ Rule failed to detect violation"
  exit 1
fi

echo "Testing exception works"
planguard -config .planguard/config.hcl -directory test/s3-public-exception/
if [ $? -eq 0 ]; then
  echo "✅ Exception applied correctly"
else
  echo "❌ Exception not working"
  exit 1
fi
```

## Debugging

### Enable Verbose Output

```bash
# See which files are being parsed
planguard -config .planguard/config.hcl -directory . 2>&1 | grep "Found"

# Output: Found 42 resources in 12 files
```

### Test Expressions

Use Terraform console to test expressions:

```bash
terraform console

> anytrue([true, false, false])
true

> jsondecode("{\"key\": \"value\"}")
{
  "key" = "value"
}
```

### Common Issues

**Issue: "No Terraform files found"**
- Check the directory path
- Ensure files have `.tf` extension
- Check exclude patterns

**Issue: "Expression evaluation error"**
- Test the expression in `terraform console`
- Check for syntax errors
- Ensure all functions are valid

**Issue: "Rule not triggering"**
- Verify resource_type matches
- Check when condition isn't filtering it out
- Test condition logic

## CI/CD Integration

### GitHub Actions

See README.md for GitHub Actions integration.

### GitLab CI

```yaml
# .gitlab-ci.yml

terraform-scan:
  stage: test
  image: golang:1.22
  before_script:
    - git clone https://github.com/your-org/planguard.git /planguard
    - cd /planguard && make build
    - export PATH=$PATH:/planguard/bin
  script:
    - cd $CI_PROJECT_DIR
    - planguard -config .planguard/config.hcl -format json > results.json
  artifacts:
    reports:
      codequality: results.json
```

### Jenkins

```groovy
pipeline {
  agent any
  stages {
    stage('Terraform Security Scan') {
      steps {
        sh '''
          planguard -config .planguard/config.hcl \
                   -directory ./terraform \
                   -format sarif > results.sarif
        '''
        recordIssues(
          tools: [sarif(pattern: 'results.sarif')]
        )
      }
    }
  }
}
```

## Best Practices

1. **Start with Default Rules**: Use shipped rules first, add custom ones as needed
2. **Use Exceptions Wisely**: Every exception should have a reason and approver
3. **Set Expiration Dates**: Time-bound exceptions for temporary situations
4. **Version Control Rules**: Keep rules in git alongside Terraform code
5. **Test Before Enforcing**: Run in warning mode first, then promote to errors
6. **Document Custom Rules**: Add remediation and references to rules
7. **Use Semantic Versioning**: Pin planguard version in CI/CD

## Performance Tips

1. **Exclude Unnecessary Paths**: Use `exclude_paths` for vendor directories
2. **Parallel Scanning**: Planguard automatically parallelizes file parsing
3. **Cache Results**: In CI, cache the planguard binary

## Getting Help

- 📖 Read the README.md
- 🐛 Check existing issues
- 💬 Start a discussion
- 📧 Contact maintainers

## Contributing

We welcome contributions! See CONTRIBUTING.md for guidelines.
