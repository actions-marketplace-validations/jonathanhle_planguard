package functions

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/planguard/planguard/pkg/config"
	"github.com/planguard/planguard/pkg/parser"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// ResourcesFunc returns resources matching a type pattern
func ResourcesFunc(ctx *parser.ScanContext) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{Name: "type", Type: cty.String},
		},
		Type: function.StaticReturnType(cty.List(cty.DynamicPseudoType)),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			resourceType := args[0].AsString()
			resources := ctx.GetResourcesByType(resourceType)
			return resourcesToCty(resources), nil
		},
	})
}

// ResourcesInFileFunc returns resources in a specific file
func ResourcesInFileFunc(ctx *parser.ScanContext) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{Name: "filepath", Type: cty.String},
		},
		Type: function.StaticReturnType(cty.List(cty.DynamicPseudoType)),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			filePath := args[0].AsString()
			resources := ctx.GetResourcesInFile(filePath)
			return resourcesToCty(resources), nil
		},
	})
}

// DayOfWeekFunc returns the current day of the week
var DayOfWeekFunc = function.New(&function.Spec{
	Params: []function.Parameter{},
	Type:   function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		day := time.Now().Weekday().String()
		return cty.StringVal(strings.ToLower(day)), nil
	},
})

// GitBranchFunc returns the current git branch
var GitBranchFunc = function.New(&function.Spec{
	Params: []function.Parameter{},
	Type:   function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		output, err := cmd.Output()
		if err != nil {
			return cty.StringVal(""), nil // Return empty string if not in git repo
		}
		return cty.StringVal(strings.TrimSpace(string(output))), nil
	},
})

// GlobMatchFunc checks if a string matches a glob pattern
var GlobMatchFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "pattern", Type: cty.String},
		{Name: "str", Type: cty.String},
	},
	Type: function.StaticReturnType(cty.Bool),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		pattern := args[0].AsString()
		str := args[1].AsString()

		matched, err := filepath.Match(pattern, str)
		if err != nil {
			return cty.NilVal, fmt.Errorf("invalid pattern: %w", err)
		}

		return cty.BoolVal(matched), nil
	},
})

// RegexMatchFunc checks if a string matches a regex pattern
var RegexMatchFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "pattern", Type: cty.String},
		{Name: "str", Type: cty.String},
	},
	Type: function.StaticReturnType(cty.Bool),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		pattern := args[0].AsString()
		str := args[1].AsString()

		matched, err := regexp.MatchString(pattern, str)
		if err != nil {
			return cty.NilVal, fmt.Errorf("invalid regex: %w", err)
		}

		return cty.BoolVal(matched), nil
	},
})

// AnyTrueFunc returns true if any element in a list is true
var AnyTrueFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "list", Type: cty.List(cty.Bool)},
	},
	Type: function.StaticReturnType(cty.Bool),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		list := args[0]

		if list.IsNull() || !list.IsKnown() {
			return cty.False, nil
		}

		if list.LengthInt() == 0 {
			return cty.False, nil
		}

		it := list.ElementIterator()
		for it.Next() {
			_, val := it.Element()
			if val.True() {
				return cty.True, nil
			}
		}

		return cty.False, nil
	},
})

// AllTrueFunc returns true if all elements in a list are true
var AllTrueFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "list", Type: cty.List(cty.Bool)},
	},
	Type: function.StaticReturnType(cty.Bool),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		list := args[0]

		if list.IsNull() || !list.IsKnown() {
			return cty.False, nil
		}

		if list.LengthInt() == 0 {
			return cty.True, nil
		}

		it := list.ElementIterator()
		for it.Next() {
			_, val := it.Element()
			if !val.True() {
				return cty.False, nil
			}
		}

		return cty.True, nil
	},
})

// HasFunc checks if an object has a specific attribute
var HasFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "object", Type: cty.DynamicPseudoType},
		{Name: "attribute", Type: cty.String},
	},
	Type: function.StaticReturnType(cty.Bool),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		obj := args[0]
		attrName := args[1].AsString()

		if obj.IsNull() || !obj.IsKnown() {
			return cty.False, nil
		}

		if !obj.Type().IsObjectType() && !obj.Type().IsMapType() {
			return cty.False, nil
		}

		if obj.Type().IsObjectType() {
			if obj.Type().HasAttribute(attrName) {
				attr := obj.GetAttr(attrName)
				// Check if the attribute exists and is not null
				return cty.BoolVal(!attr.IsNull()), nil
			}
			return cty.False, nil
		}

		if obj.Type().IsMapType() {
			return cty.BoolVal(obj.HasIndex(cty.StringVal(attrName)).True()), nil
		}

		return cty.False, nil
	},
})

// Helper function to convert resources to cty values
func resourcesToCty(resources []*config.Resource) cty.Value {
	if len(resources) == 0 {
		return cty.ListValEmpty(cty.DynamicPseudoType)
	}

	vals := make([]cty.Value, len(resources))
	for i, resource := range resources {
		vals[i] = resourceToCty(resource)
	}

	return cty.ListVal(vals)
}

// Helper function to convert a single resource to cty value
func resourceToCty(resource *config.Resource) cty.Value {
	attrs := make(map[string]cty.Value)

	// Add metadata
	attrs["type"] = cty.StringVal(resource.Type)
	attrs["name"] = cty.StringVal(resource.Name)
	attrs["file"] = cty.StringVal(resource.File)
	attrs["line"] = cty.NumberIntVal(int64(resource.Line))

	// Add resource attributes
	for key, val := range resource.Attributes {
		attrs[key] = val
	}

	return cty.ObjectVal(attrs)
}
