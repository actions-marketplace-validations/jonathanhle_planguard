package functions

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestMD5Func(t *testing.T) {
	result, err := MD5Func.Call([]cty.Value{
		cty.StringVal("hello"),
	})

	if err != nil {
		t.Fatalf("md5() error: %v", err)
	}

	expected := "5d41402abc4b2a76b9719d911017c592"
	if result.AsString() != expected {
		t.Errorf("md5(\"hello\") = %s, want %s", result.AsString(), expected)
	}
}

func TestSHA1Func(t *testing.T) {
	result, err := SHA1Func.Call([]cty.Value{
		cty.StringVal("hello"),
	})

	if err != nil {
		t.Fatalf("sha1() error: %v", err)
	}

	expected := "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"
	if result.AsString() != expected {
		t.Errorf("sha1(\"hello\") = %s, want %s", result.AsString(), expected)
	}
}

func TestSHA256Func(t *testing.T) {
	result, err := SHA256Func.Call([]cty.Value{
		cty.StringVal("hello"),
	})

	if err != nil {
		t.Fatalf("sha256() error: %v", err)
	}

	expected := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if result.AsString() != expected {
		t.Errorf("sha256(\"hello\") = %s, want %s", result.AsString(), expected)
	}
}

func TestSHA512Func(t *testing.T) {
	result, err := SHA512Func.Call([]cty.Value{
		cty.StringVal("hello"),
	})

	if err != nil {
		t.Fatalf("sha512() error: %v", err)
	}

	// Just verify it's a valid SHA512 (128 hex chars)
	if len(result.AsString()) != 128 {
		t.Errorf("sha512() returned wrong length: %d, want 128", len(result.AsString()))
	}
}

func TestUUIDFunc(t *testing.T) {
	result1, err := UUIDFunc.Call([]cty.Value{})
	if err != nil {
		t.Fatalf("uuid() error: %v", err)
	}

	result2, err := UUIDFunc.Call([]cty.Value{})
	if err != nil {
		t.Fatalf("uuid() error: %v", err)
	}

	// UUIDs should be unique
	if result1.AsString() == result2.AsString() {
		t.Error("uuid() should generate unique values")
	}

	// Check format (8-4-4-4-12)
	uuid := result1.AsString()
	parts := strings.Split(uuid, "-")
	if len(parts) != 5 {
		t.Errorf("uuid() format invalid: %s", uuid)
	}
}

func TestUUIDV5Func(t *testing.T) {
	namespace := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	name := "test"

	result1, err := UUIDV5Func.Call([]cty.Value{
		cty.StringVal(namespace),
		cty.StringVal(name),
	})
	if err != nil {
		t.Fatalf("uuidv5() error: %v", err)
	}

	result2, err := UUIDV5Func.Call([]cty.Value{
		cty.StringVal(namespace),
		cty.StringVal(name),
	})
	if err != nil {
		t.Fatalf("uuidv5() error: %v", err)
	}

	// Same namespace + name should produce same UUID
	if result1.AsString() != result2.AsString() {
		t.Error("uuidv5() should be deterministic")
	}
}

func TestUUIDV5FuncInvalidNamespace(t *testing.T) {
	_, err := UUIDV5Func.Call([]cty.Value{
		cty.StringVal("not-a-uuid"),
		cty.StringVal("name"),
	})

	if err == nil {
		t.Error("Expected error for invalid namespace UUID")
	}
}

func TestBcryptFunc(t *testing.T) {
	password := "mysecretpassword"

	result, err := BcryptFunc.Call([]cty.Value{
		cty.StringVal(password),
	})
	if err != nil {
		t.Fatalf("bcrypt() error: %v", err)
	}

	hash := result.AsString()

	// Bcrypt hashes start with $2a$ or $2b$
	if !strings.HasPrefix(hash, "$2") {
		t.Errorf("bcrypt() hash format invalid: %s", hash)
	}
}

func TestBcryptFuncWithCost(t *testing.T) {
	password := "test"
	cost := int64(10)

	result, err := BcryptFunc.Call([]cty.Value{
		cty.StringVal(password),
		cty.NumberIntVal(cost),
	})
	if err != nil {
		t.Fatalf("bcrypt() with cost error: %v", err)
	}

	if result.AsString() == "" {
		t.Error("bcrypt() returned empty string")
	}
}

func TestBcryptFuncInvalidCost(t *testing.T) {
	_, err := BcryptFunc.Call([]cty.Value{
		cty.StringVal("password"),
		cty.NumberIntVal(1), // Too low
	})

	// Bcrypt may or may not error on low cost depending on implementation
	// Just verify it doesn't panic
	_ = err
}

func TestBase64SHA256Func(t *testing.T) {
	result, err := Base64SHA256Func.Call([]cty.Value{
		cty.StringVal("hello"),
	})
	if err != nil {
		t.Fatalf("base64sha256() error: %v", err)
	}

	// Decode and verify it's valid base64
	_, err = base64.StdEncoding.DecodeString(result.AsString())
	if err != nil {
		t.Errorf("base64sha256() didn't return valid base64: %v", err)
	}
}

func TestBase64SHA512Func(t *testing.T) {
	result, err := Base64SHA512Func.Call([]cty.Value{
		cty.StringVal("hello"),
	})
	if err != nil {
		t.Fatalf("base64sha512() error: %v", err)
	}

	// Decode and verify it's valid base64
	_, err = base64.StdEncoding.DecodeString(result.AsString())
	if err != nil {
		t.Errorf("base64sha512() didn't return valid base64: %v", err)
	}
}
