# Common Tagging Rules

rule "require_tags" {
  name     = "Resources should have required tags"
  severity = "warning"
  
  resource_type = "aws_*"
  
  condition {
    expression = <<-EXPR
      !has(self, "tags") ||
      !has(try(self.tags, {}), "Environment") ||
      !has(try(self.tags, {}), "Owner")
    EXPR
  }
  
  message = "Resources should have Environment and Owner tags"
}

rule "environment_tag_values" {
  name     = "Environment tag should have valid values"
  severity = "warning"
  
  resource_type = "aws_*"
  
  when {
    expression = "has(self, \"tags\") && has(try(self.tags, {}), \"Environment\")"
  }
  
  condition {
    expression = <<-EXPR
      !contains(["dev", "staging", "prod", "test"], lower(self.tags.Environment))
    EXPR
  }
  
  message = "Environment tag must be one of: dev, staging, prod, test"
}
