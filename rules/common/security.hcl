# Security Rules - Dangerous Terraform Patterns

# Dangerous Functions

rule "dangerous_nonsensitive_function" {
  name     = "Prevent nonsensitive() function usage"
  severity = "error"

  resource_type = "*"

  condition {
    expression = "contains_function_call(\"nonsensitive\")"
  }

  message = "The 'nonsensitive()' function can expose sensitive data and is blocked for security reasons"

  remediation = <<-EOT
    Remove the nonsensitive() function call:

    # DON'T:
    output "secret" {
      value = nonsensitive(var.sensitive_data)
    }

    # DO: Keep sensitive data marked as sensitive
    output "secret" {
      value     = var.sensitive_data
      sensitive = true
    }

    The nonsensitive() function removes the sensitive marking from values,
    which can lead to accidental exposure of secrets in logs and outputs.
  EOT
}

# Dangerous Data Sources
# These data sources can perform risky operations and should be blocked

rule "dangerous_http_data_source" {
  name     = "Prevent http data source usage"
  severity = "error"

  resource_type = "http"

  condition {
    expression = "true"
  }

  message = "The 'http' data source can make arbitrary HTTP requests and is blocked for security reasons"

  remediation = <<-EOT
    Remove the http data source and use a safer alternative:

    # DON'T:
    data "http" "example" {
      url = "https://api.example.com/data"
    }

    # DO: Use pre-fetched data, terraform_remote_state, or approved data sources
    data "terraform_remote_state" "example" {
      backend = "s3"
      config = {
        bucket = "terraform-state"
        key    = "data.tfstate"
      }
    }
  EOT
}

rule "dangerous_dns_data_source" {
  name     = "Prevent dns data source usage"
  severity = "error"

  resource_type = "dns"

  condition {
    expression = "true"
  }

  message = "The 'dns' data source can make DNS queries and is blocked for security reasons"

  remediation = <<-EOT
    Remove the dns data source:

    # DON'T:
    data "dns" "example" {
      name = "example.com"
    }

    # DO: Hard-code known DNS values or use approved data sources
  EOT
}

rule "dangerous_external_data_source" {
  name     = "Prevent external data source usage"
  severity = "error"

  resource_type = "external"

  condition {
    expression = "true"
  }

  message = "The 'external' data source can execute arbitrary external programs and is blocked for security reasons"

  remediation = <<-EOT
    Remove the external data source and use a safer alternative:

    # DON'T:
    data "external" "example" {
      program = ["python", "script.py"]
    }

    # DO: Use Terraform providers, terraform_remote_state, or pre-computed values
    # If you truly need external data, work with your security team to create
    # an approved Terraform provider instead.
  EOT
}

