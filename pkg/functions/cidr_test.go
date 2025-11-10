package functions

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestCIDRHostFunc(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		hostnum  int64
		expected string
		wantErr  bool
	}{
		{
			name:     "basic /24",
			prefix:   "192.168.1.0/24",
			hostnum:  5,
			expected: "192.168.1.5",
		},
		{
			name:     "first host",
			prefix:   "10.0.0.0/16",
			hostnum:  1,
			expected: "10.0.0.1",
		},
		{
			name:    "invalid CIDR",
			prefix:  "invalid",
			hostnum: 1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CIDRHostFunc.Call([]cty.Value{
				cty.StringVal(tt.prefix),
				cty.NumberIntVal(tt.hostnum),
			})

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.AsString() != tt.expected {
				t.Errorf("cidrhost() = %s, want %s", result.AsString(), tt.expected)
			}
		})
	}
}

func TestCIDRNetmaskFunc(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		expected string
		wantErr  bool
	}{
		{
			name:     "/24 network",
			prefix:   "192.168.1.0/24",
			expected: "255.255.255.0",
		},
		{
			name:     "/16 network",
			prefix:   "10.0.0.0/16",
			expected: "255.255.0.0",
		},
		{
			name:     "/8 network",
			prefix:   "172.0.0.0/8",
			expected: "255.0.0.0",
		},
		{
			name:    "invalid CIDR",
			prefix:  "not-a-cidr",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CIDRNetmaskFunc.Call([]cty.Value{
				cty.StringVal(tt.prefix),
			})

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.AsString() != tt.expected {
				t.Errorf("cidrnetmask() = %s, want %s", result.AsString(), tt.expected)
			}
		})
	}
}

func TestCIDRSubnetFunc(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		newbits  int64
		netnum   int64
		expected string
		wantErr  bool
	}{
		{
			name:     "split /24 into /26",
			prefix:   "192.168.1.0/24",
			newbits:  2,
			netnum:   1,
			expected: "192.168.1.64/26",
		},
		{
			name:    "invalid CIDR",
			prefix:  "invalid",
			newbits: 2,
			netnum:  1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CIDRSubnetFunc.Call([]cty.Value{
				cty.StringVal(tt.prefix),
				cty.NumberIntVal(tt.newbits),
				cty.NumberIntVal(tt.netnum),
			})

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.AsString() != tt.expected {
				t.Errorf("cidrsubnet() = %s, want %s", result.AsString(), tt.expected)
			}
		})
	}
}
