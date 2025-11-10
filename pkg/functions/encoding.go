package functions

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/url"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// Base64EncodeFunc encodes a string to base64
var Base64EncodeFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "str", Type: cty.String},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		str := args[0].AsString()
		encoded := base64.StdEncoding.EncodeToString([]byte(str))
		return cty.StringVal(encoded), nil
	},
})

// Base64DecodeFunc decodes a base64 string
var Base64DecodeFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "str", Type: cty.String},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		str := args[0].AsString()
		decoded, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			return cty.NilVal, fmt.Errorf("invalid base64: %w", err)
		}
		return cty.StringVal(string(decoded)), nil
	},
})

// Base64GzipFunc compresses and base64 encodes a string
var Base64GzipFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "str", Type: cty.String},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		str := args[0].AsString()

		var buf bytes.Buffer
		gzipWriter := gzip.NewWriter(&buf)
		_, err := gzipWriter.Write([]byte(str))
		if err != nil {
			return cty.NilVal, fmt.Errorf("gzip compression failed: %w", err)
		}
		err = gzipWriter.Close()
		if err != nil {
			return cty.NilVal, fmt.Errorf("gzip close failed: %w", err)
		}

		encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
		return cty.StringVal(encoded), nil
	},
})

// URLEncodeFunc URL encodes a string
var URLEncodeFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "str", Type: cty.String},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		str := args[0].AsString()
		encoded := url.QueryEscape(str)
		return cty.StringVal(encoded), nil
	},
})
