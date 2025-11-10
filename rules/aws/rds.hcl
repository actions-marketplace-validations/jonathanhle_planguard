# AWS RDS Security Rules

rule "aws_rds_encryption" {
  name     = "RDS instances must have encryption enabled"
  severity = "error"
  
  resource_type = "aws_db_instance"
  
  condition {
    expression = "!has(self, \"storage_encrypted\") || try(self.storage_encrypted, false) != true"
  }
  
  message = "RDS instances must have storage encryption enabled for compliance"
  
  remediation = <<-EOT
    Enable encryption on the RDS instance:
    
    resource "aws_db_instance" "example" {
      storage_encrypted = true
      # ... other config
    }
  EOT
}

rule "aws_rds_public_access" {
  name     = "RDS instances should not be publicly accessible"
  severity = "error"
  
  resource_type = "aws_db_instance"
  
  condition {
    expression = "has(self, \"publicly_accessible\") && try(self.publicly_accessible, false) == true"
  }
  
  message = "RDS instances must not be publicly accessible"
}

rule "aws_rds_backup_retention" {
  name     = "RDS instances should have adequate backup retention"
  severity = "warning"
  
  resource_type = "aws_db_instance"
  
  condition {
    expression = "!has(self, \"backup_retention_period\") || tonumber(try(self.backup_retention_period, \"0\")) < 7"
  }
  
  message = "RDS instances should have at least 7 days backup retention"
}
