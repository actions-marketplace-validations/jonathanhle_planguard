package functions

import (
	"encoding/json"
	"fmt"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

// JSONDecodeFunc decodes a JSON string into a cty value
var JSONDecodeFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "str",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.DynamicPseudoType),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		jsonStr := args[0].AsString()

		var raw interface{}
		err := json.Unmarshal([]byte(jsonStr), &raw)
		if err != nil {
			return cty.NilVal, fmt.Errorf("invalid JSON: %w", err)
		}

		// Convert JSON to cty value
		return jsonToCty(raw), nil
	},
})

// JSONEncodeFunc encodes a cty value as JSON
var JSONEncodeFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "val",
			Type: cty.DynamicPseudoType,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		val := args[0]

		// Use cty's JSON encoder
		jsonBytes, err := ctyjson.Marshal(val, val.Type())
		if err != nil {
			return cty.NilVal, fmt.Errorf("failed to encode JSON: %w", err)
		}

		return cty.StringVal(string(jsonBytes)), nil
	},
})

// jsonToCty converts a Go value from JSON into a cty.Value
func jsonToCty(val interface{}) cty.Value {
	switch v := val.(type) {
	case nil:
		return cty.NullVal(cty.DynamicPseudoType)
	case bool:
		return cty.BoolVal(v)
	case float64:
		return cty.NumberFloatVal(v)
	case string:
		return cty.StringVal(v)
	case []interface{}:
		if len(v) == 0 {
			return cty.ListValEmpty(cty.DynamicPseudoType)
		}
		vals := make([]cty.Value, len(v))
		for i, item := range v {
			vals[i] = jsonToCty(item)
		}
		return cty.ListVal(vals)
	case map[string]interface{}:
		if len(v) == 0 {
			return cty.MapValEmpty(cty.DynamicPseudoType)
		}
		vals := make(map[string]cty.Value)
		for key, item := range v {
			vals[key] = jsonToCty(item)
		}
		return cty.MapVal(vals)
	default:
		return cty.NullVal(cty.DynamicPseudoType)
	}
}
