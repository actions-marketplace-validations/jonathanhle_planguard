# Example Terraform configuration with security issues

resource "aws_s3_bucket" "public_bucket" {
  bucket = "my-public-bucket"
  acl    = "public-read"  # This will trigger aws_s3_public_read rule
  
  tags = {
    Name = "Public Bucket"
    # Missing Environment and Owner tags - will trigger require_tags rule
  }
}

resource "aws_db_instance" "database" {
  allocated_storage   = 20
  engine              = "mysql"
  instance_class      = "db.t3.micro"
  # storage_encrypted is not set - will trigger aws_rds_encryption rule
  publicly_accessible = true  # Will trigger aws_rds_public_access rule
  
  tags = {
    Environment = "dev"
    Owner       = "team@example.com"
  }
}

resource "aws_security_group" "allow_all" {
  name        = "allow_all"
  description = "Allow all inbound traffic"
  
  ingress {
    from_port   = 0
    to_port     = 65535
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]  # Will trigger aws_security_group_ingress_all rule
  }
  
  tags = {
    Environment = "prod"
    Owner       = "security-team"
  }
}

resource "aws_instance" "web" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"
  
  # Missing metadata_options - will trigger aws_ec2_imdsv2 rule
  
  tags = {
    Environment = "prod"
    Owner       = "devops"
  }
}
