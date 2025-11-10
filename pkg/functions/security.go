package functions

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/planguard/planguard/pkg/parser"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// ContainsFunctionCallFunc checks if any expression in the current resource contains a specific function call
func ContainsFunctionCallFunc(ctx *parser.ScanContext) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{Name: "function_name", Type: cty.String},
		},
		Type: function.StaticReturnType(cty.Bool),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			targetFunc := args[0].AsString()

			if ctx.CurrentResource == nil {
				return cty.False, nil
			}

			// Check all raw expressions in the current resource
			for _, expr := range ctx.CurrentResource.RawExprs {
				if containsFunctionCall(expr, targetFunc) {
					return cty.True, nil
				}
			}

			return cty.False, nil
		},
	})
}

// containsFunctionCall recursively walks an HCL expression to find function calls
func containsFunctionCall(expr hcl.Expression, targetFunc string) bool {
	if expr == nil {
		return false
	}

	// Walk the expression AST
	return walkExpression(expr, targetFunc)
}

// walkExpression recursively traverses the expression tree looking for function calls
func walkExpression(expr hcl.Expression, targetFunc string) bool {
	switch e := expr.(type) {
	case *hclsyntax.FunctionCallExpr:
		// Direct function call - check if it matches
		if e.Name == targetFunc {
			return true
		}
		// Check arguments
		for _, arg := range e.Args {
			if walkExpression(arg, targetFunc) {
				return true
			}
		}

	case *hclsyntax.TemplateExpr:
		// Template expression - check all parts
		for _, part := range e.Parts {
			if walkExpression(part, targetFunc) {
				return true
			}
		}

	case *hclsyntax.TemplateWrapExpr:
		// Wrapped expression
		return walkExpression(e.Wrapped, targetFunc)

	case *hclsyntax.ConditionalExpr:
		// Conditional expression - check all branches
		return walkExpression(e.Condition, targetFunc) ||
			walkExpression(e.TrueResult, targetFunc) ||
			walkExpression(e.FalseResult, targetFunc)

	case *hclsyntax.BinaryOpExpr:
		// Binary operation - check both sides
		return walkExpression(e.LHS, targetFunc) ||
			walkExpression(e.RHS, targetFunc)

	case *hclsyntax.UnaryOpExpr:
		// Unary operation - check the value
		return walkExpression(e.Val, targetFunc)

	case *hclsyntax.ParenthesesExpr:
		// Parentheses - check inner expression
		return walkExpression(e.Expression, targetFunc)

	case *hclsyntax.IndexExpr:
		// Index expression - check collection and key
		return walkExpression(e.Collection, targetFunc) ||
			walkExpression(e.Key, targetFunc)

	case *hclsyntax.RelativeTraversalExpr:
		// Traversal - check source
		return walkExpression(e.Source, targetFunc)

	case *hclsyntax.SplatExpr:
		// Splat expression - check source and each
		result := walkExpression(e.Source, targetFunc)
		if e.Each != nil {
			result = result || walkExpression(e.Each, targetFunc)
		}
		return result

	case *hclsyntax.ForExpr:
		// For expression - check all parts
		result := walkExpression(e.CollExpr, targetFunc)
		if e.KeyExpr != nil {
			result = result || walkExpression(e.KeyExpr, targetFunc)
		}
		if e.ValExpr != nil {
			result = result || walkExpression(e.ValExpr, targetFunc)
		}
		if e.CondExpr != nil {
			result = result || walkExpression(e.CondExpr, targetFunc)
		}
		return result

	case *hclsyntax.ObjectConsExpr:
		// Object construction - check all items
		for _, item := range e.Items {
			if walkExpression(item.KeyExpr, targetFunc) || walkExpression(item.ValueExpr, targetFunc) {
				return true
			}
		}

	case *hclsyntax.TupleConsExpr:
		// Tuple construction - check all expressions
		for _, expr := range e.Exprs {
			if walkExpression(expr, targetFunc) {
				return true
			}
		}
	}

	// Not found in this expression
	return false
}
