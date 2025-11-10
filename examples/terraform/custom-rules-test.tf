# Test file for custom organizational rules
# This file demonstrates various custom rule violations

# Should trigger: sox_segregation_of_duties
# Admin IAM role mixed with other resources in same file
resource "aws_iam_role" "admin_role" {
  name = "my-admin-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.amazonaws.com"
      }
    }]
  })
}

# This S3 bucket in same file violates segregation of duties
resource "aws_s3_bucket" "mixed_file" {
  bucket = "test-bucket-mixed"
}

# Should trigger: expensive_instances_require_approval
# Large instance without CostApproval tag
resource "aws_instance" "large_instance" {
  ami           = "ami-12345678"
  instance_type = "t3.2xlarge"

  tags = {
    Name        = "large-instance"
    Environment = "production"
    # Missing CostApproval tag
  }
}

# Should trigger: resource_naming_convention
# Resource name has uppercase letters
resource "aws_s3_bucket" "BadNaming" {
  bucket = "test-bucket-bad-naming"
}

# Should NOT trigger: lowercase with hyphens is correct
resource "aws_s3_bucket" "good-naming-example" {
  bucket = "test-bucket-good"
}

# Should trigger: prod_requires_backup
# Production database with insufficient backup retention
resource "aws_db_instance" "prod_db" {
  allocated_storage       = 20
  engine                  = "mysql"
  instance_class          = "db.t3.micro"
  backup_retention_period = 3 # Less than 7 days

  tags = {
    Environment = "prod"
  }
}

# Should NOT trigger: non-prod database
resource "aws_db_instance" "dev_db" {
  allocated_storage       = 20
  engine                  = "mysql"
  instance_class          = "db.t3.micro"
  backup_retention_period = 1

  tags = {
    Environment = "dev"
  }
}
