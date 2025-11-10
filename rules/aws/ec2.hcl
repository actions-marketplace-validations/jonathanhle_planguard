# AWS EC2 Security Rules

rule "aws_ec2_imdsv2" {
  name     = "EC2 instances should use IMDSv2"
  severity = "warning"
  
  resource_type = "aws_instance"
  
  condition {
    expression = <<-EXPR
      !has(self, "metadata_options") ||
      !has(try(self.metadata_options, {}), "http_tokens") ||
      try(self.metadata_options.http_tokens, "") != "required"
    EXPR
  }
  
  message = "EC2 instances should require IMDSv2 for enhanced security"
}

rule "aws_security_group_ingress_all" {
  name     = "Security groups should not allow ingress from 0.0.0.0/0"
  severity = "error"
  
  resource_type = "aws_security_group"
  
  condition {
    expression = <<-EXPR
      anytrue([
        for rule in try(self.ingress, []) :
        contains(try(rule.cidr_blocks, []), "0.0.0.0/0")
      ])
    EXPR
  }
  
  message = "Security groups must not allow ingress from 0.0.0.0/0"
}
