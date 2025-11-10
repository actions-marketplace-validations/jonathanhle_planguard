# Example Custom Organizational Rules
# These demonstrate how to write custom rules for your organization

# Cost Control - Large instances require approval
rule "expensive_instances_require_approval" {
  name     = "Expensive instances require approval tag"
  severity = "error"

  resource_type = "aws_instance"

  when {
    expression = <<-EXPR
      can(regex(".*2xlarge.*", try(self.instance_type, ""))) ||
      can(regex(".*4xlarge.*", try(self.instance_type, ""))) ||
      can(regex(".*8xlarge.*", try(self.instance_type, ""))) ||
      can(regex(".*16xlarge.*", try(self.instance_type, "")))
    EXPR
  }

  condition {
    expression = "!has(try(self.tags, {}), \"CostApproval\")"
  }

  message = "Large instances (2xlarge+) require CostApproval tag with approver email"

  remediation = <<-EOT
    Add CostApproval tag to large instance:

    resource "aws_instance" "example" {
      instance_type = "t3.2xlarge"

      tags = {
        CostApproval = "finance-team@example.com"
        # ... other tags
      }
    }
  EOT
}

# Naming Conventions
rule "resource_naming_convention" {
  name     = "Resources must follow naming convention"
  severity = "warning"

  resource_type = "aws_*"

  condition {
    expression = "!regex_match(\"^[a-z][a-z0-9_-]*[a-z0-9]$\", self.name)"
  }

  message      = "Resource names must be lowercase, alphanumeric with underscores/hyphens, starting with letter"
  remediation  = "Rename resource to use lowercase letters, numbers, underscores, and hyphens only"
}

# Production Safeguards - Backup retention
rule "prod_requires_backup" {
  name     = "Production databases need backup strategy"
  severity = "error"

  resource_type = "aws_db_instance"

  when {
    expression = "lookup(try(self.tags, {}), \"Environment\", \"\") == \"prod\""
  }

  condition {
    expression = <<-EXPR
      !has(self, "backup_retention_period") ||
      tonumber(try(self.backup_retention_period, 0)) < 7
    EXPR
  }

  message = "Production databases must have at least 7-day backup retention"

  remediation = <<-EOT
    Set backup retention for production database:

    resource "aws_db_instance" "prod" {
      backup_retention_period = 7  # At least 7 days

      tags = {
        Environment = "prod"
      }
    }
  EOT
}

# Friday Deployment Freeze
rule "no_prod_friday" {
  name     = "No production changes on Fridays"
  severity = "error"

  resource_type = "*"

  when {
    expression = <<-EXPR
      day_of_week() == "friday" &&
      can(regex(".*/prod/.*", self.file))
    EXPR
  }

  condition {
    expression = "true"
  }

  message = "Production changes not allowed on Fridays per change control policy"

  remediation = "Schedule this deployment for Monday through Thursday, or request emergency change approval"
}

# IAM Admin Role Naming
rule "iam_admin_naming" {
  name     = "Admin IAM roles must follow naming pattern"
  severity = "warning"

  resource_type = "aws_iam_role"

  when {
    expression = "can(regex(\".*admin.*\", lower(try(self.name, \"\"))))"
  }

  condition {
    expression = "!regex_match(\"^admin-[a-z0-9-]+$\", try(self.name, \"\"))"
  }

  message = "Admin IAM roles must be named 'admin-<purpose>' (lowercase with hyphens)"

  remediation = <<-EOT
    Rename admin role to follow pattern:

    # Good examples:
    resource "aws_iam_role" "admin-ec2" { ... }
    resource "aws_iam_role" "admin-s3-readonly" { ... }

    # Bad examples:
    resource "aws_iam_role" "AdminRole" { ... }
    resource "aws_iam_role" "my_admin_role" { ... }
  EOT
}
