package functions

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"golang.org/x/crypto/bcrypt"
)

// MD5Func computes MD5 hash
var MD5Func = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "str", Type: cty.String},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		str := args[0].AsString()
		hash := md5.Sum([]byte(str))
		return cty.StringVal(hex.EncodeToString(hash[:])), nil
	},
})

// SHA1Func computes SHA1 hash
var SHA1Func = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "str", Type: cty.String},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		str := args[0].AsString()
		hash := sha1.Sum([]byte(str))
		return cty.StringVal(hex.EncodeToString(hash[:])), nil
	},
})

// SHA256Func computes SHA256 hash
var SHA256Func = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "str", Type: cty.String},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		str := args[0].AsString()
		hash := sha256.Sum256([]byte(str))
		return cty.StringVal(hex.EncodeToString(hash[:])), nil
	},
})

// SHA512Func computes SHA512 hash
var SHA512Func = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "str", Type: cty.String},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		str := args[0].AsString()
		hash := sha512.Sum512([]byte(str))
		return cty.StringVal(hex.EncodeToString(hash[:])), nil
	},
})

// Base64SHA256Func computes SHA256 and base64 encodes it
var Base64SHA256Func = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "str", Type: cty.String},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		str := args[0].AsString()
		hash := sha256.Sum256([]byte(str))
		encoded := base64.StdEncoding.EncodeToString(hash[:])
		return cty.StringVal(encoded), nil
	},
})

// Base64SHA512Func computes SHA512 and base64 encodes it
var Base64SHA512Func = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "str", Type: cty.String},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		str := args[0].AsString()
		hash := sha512.Sum512([]byte(str))
		encoded := base64.StdEncoding.EncodeToString(hash[:])
		return cty.StringVal(encoded), nil
	},
})

// BcryptFunc generates a bcrypt hash
var BcryptFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "str", Type: cty.String},
	},
	VarParam: &function.Parameter{
		Name: "cost",
		Type: cty.Number,
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		str := args[0].AsString()
		cost := bcrypt.DefaultCost

		if len(args) > 1 {
			costFloat, _ := args[1].AsBigFloat().Float64()
			cost = int(costFloat)
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(str), cost)
		if err != nil {
			return cty.NilVal, fmt.Errorf("bcrypt failed: %w", err)
		}

		return cty.StringVal(string(hash)), nil
	},
})

// UUIDFunc generates a random UUID
var UUIDFunc = function.New(&function.Spec{
	Params: []function.Parameter{},
	Type:   function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		return cty.StringVal(uuid.New().String()), nil
	},
})

// UUIDV5Func generates a UUID v5
var UUIDV5Func = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "namespace", Type: cty.String},
		{Name: "name", Type: cty.String},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		ns := args[0].AsString()
		name := args[1].AsString()

		// Parse namespace UUID
		namespaceUUID, err := uuid.Parse(ns)
		if err != nil {
			return cty.NilVal, fmt.Errorf("invalid namespace UUID: %w", err)
		}

		id := uuid.NewSHA1(namespaceUUID, []byte(name))
		return cty.StringVal(id.String()), nil
	},
})
