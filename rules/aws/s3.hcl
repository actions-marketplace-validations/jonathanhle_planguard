# AWS S3 Security Rules

rule "aws_s3_public_read" {
  name     = "Prevent public-read S3 buckets"
  severity = "error"

  resource_type = "aws_s3_bucket"

  condition {
    expression = "try(self.acl, \"\") == \"public-read\""
  }

  message = "S3 buckets must not have public-read ACL"
  
  remediation = <<-EOT
    Remove the public-read ACL:
    
    resource "aws_s3_bucket" "example" {
      bucket = "my-bucket"
      # Remove: acl = "public-read"
    }
  EOT
}

rule "aws_s3_public_readwrite" {
  name     = "Prevent public-read-write S3 buckets"
  severity = "error"

  resource_type = "aws_s3_bucket"

  condition {
    expression = "try(self.acl, \"\") == \"public-read-write\""
  }

  message = "S3 buckets must not have public-read-write ACL"
}

rule "aws_s3_versioning" {
  name     = "S3 buckets should have versioning enabled"
  severity = "warning"

  resource_type = "aws_s3_bucket"

  condition {
    expression = "!has(self, \"versioning\") || try(self.versioning.enabled, false) != true"
  }

  message = "S3 buckets should enable versioning for data protection"
}
