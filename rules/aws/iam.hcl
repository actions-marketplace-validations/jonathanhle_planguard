# AWS IAM Security Rules

rule "aws_iam_wildcard_actions" {
  name     = "Prevent wildcard IAM actions with wildcard resources"
  severity = "error"
  
  resource_type = "aws_iam_policy"
  
  condition {
    expression = <<-EXPR
      anytrue([
        for stmt in jsondecode(self.policy).Statement :
        contains(try(tolist(stmt.Action), [stmt.Action]), "*") &&
        contains(try(tolist(stmt.Resource), [stmt.Resource]), "*")
      ])
    EXPR
  }
  
  message = "IAM policies must not use wildcard (*) for both Action and Resource"
}

rule "aws_iam_admin_policy" {
  name     = "Avoid overly permissive IAM policies"
  severity = "warning"
  
  resource_type = "aws_iam_policy"
  
  condition {
    expression = <<-EXPR
      anytrue([
        for stmt in jsondecode(self.policy).Statement :
        contains(try(tolist(stmt.Action), [stmt.Action]), "*:*")
      ])
    EXPR
  }
  
  message = "IAM policy grants excessive permissions with *:* action"
}
