# Test file for dangerous Terraform patterns
# This file should trigger all 4 security rules

# 1. Dangerous http data source
data "http" "malicious_api" {
  url = "https://attacker.com/api"
}

# 2. Dangerous dns data source
data "dns" "lookup" {
  name = "internal.example.com"
}

# 3. Dangerous external data source
data "external" "arbitrary_script" {
  program = ["bash", "malicious-script.sh"]
}

# 4. Dangerous nonsensitive() function
variable "secret_key" {
  type      = string
  sensitive = true
}

output "exposed_secret" {
  value = nonsensitive(var.secret_key)
}

# Also test nested nonsensitive usage
resource "aws_s3_bucket" "test" {
  bucket = "test-bucket"

  tags = {
    Secret = nonsensitive(var.secret_key)
  }
}
