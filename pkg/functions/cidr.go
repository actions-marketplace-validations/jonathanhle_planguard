package functions

import (
	"fmt"
	"math/big"
	"net"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// CIDRHostFunc calculates a host IP within a CIDR block
var CIDRHostFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "prefix", Type: cty.String},
		{Name: "hostnum", Type: cty.Number},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		prefix := args[0].AsString()
		hostnum := args[1].AsBigFloat()

		_, network, err := net.ParseCIDR(prefix)
		if err != nil {
			return cty.NilVal, fmt.Errorf("invalid CIDR: %w", err)
		}

		hostInt, _ := hostnum.Int64()
		ip := cidrHost(network.IP, network.Mask, int(hostInt))

		return cty.StringVal(ip.String()), nil
	},
})

// CIDRNetmaskFunc returns the netmask of a CIDR block
var CIDRNetmaskFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "prefix", Type: cty.String},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		prefix := args[0].AsString()

		_, network, err := net.ParseCIDR(prefix)
		if err != nil {
			return cty.NilVal, fmt.Errorf("invalid CIDR: %w", err)
		}

		mask := net.IP(network.Mask)
		return cty.StringVal(mask.String()), nil
	},
})

// CIDRSubnetFunc calculates a subnet address within a CIDR block
var CIDRSubnetFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "prefix", Type: cty.String},
		{Name: "newbits", Type: cty.Number},
		{Name: "netnum", Type: cty.Number},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		prefix := args[0].AsString()
		newbits := args[1].AsBigFloat()
		netnum := args[2].AsBigFloat()

		_, network, err := net.ParseCIDR(prefix)
		if err != nil {
			return cty.NilVal, fmt.Errorf("invalid CIDR: %w", err)
		}

		newbitsInt, _ := newbits.Int64()
		netnumInt, _ := netnum.Int64()

		subnet := cidrSubnet(network, int(newbitsInt), int(netnumInt))
		return cty.StringVal(subnet.String()), nil
	},
})

// CIDRSubnetsFunc calculates multiple subnet addresses
var CIDRSubnetsFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "prefix", Type: cty.String},
	},
	VarParam: &function.Parameter{
		Name: "newbits",
		Type: cty.Number,
	},
	Type: function.StaticReturnType(cty.List(cty.String)),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		prefix := args[0].AsString()

		_, network, err := net.ParseCIDR(prefix)
		if err != nil {
			return cty.NilVal, fmt.Errorf("invalid CIDR: %w", err)
		}

		var subnets []cty.Value
		for i := 1; i < len(args); i++ {
			newbits := args[i].AsBigFloat()
			newbitsInt, _ := newbits.Int64()

			subnet := cidrSubnet(network, int(newbitsInt), i-1)
			subnets = append(subnets, cty.StringVal(subnet.String()))
		}

		return cty.ListVal(subnets), nil
	},
})

// Helper functions for CIDR calculations

func cidrHost(base net.IP, mask net.IPMask, hostNum int) net.IP {
	ip := make(net.IP, len(base))
	copy(ip, base)

	// Convert hostNum to big.Int
	hostBig := big.NewInt(int64(hostNum))

	// Add to IP
	for i := len(ip) - 1; i >= 0 && hostBig.Sign() > 0; i-- {
		carry := new(big.Int).And(hostBig, big.NewInt(0xFF))
		ip[i] = byte(carry.Int64()) + ip[i]
		hostBig.Rsh(hostBig, 8)
	}

	return ip
}

func cidrSubnet(network *net.IPNet, newbits, netnum int) *net.IPNet {
	ones, bits := network.Mask.Size()
	newOnes := ones + newbits

	if newOnes > bits {
		return network
	}

	newMask := net.CIDRMask(newOnes, bits)
	base := network.IP.Mask(network.Mask)

	// Calculate subnet base
	shiftAmount := uint(bits - newOnes)
	subnet := cidrHost(base, newMask, netnum<<shiftAmount)

	return &net.IPNet{
		IP:   subnet.Mask(newMask),
		Mask: newMask,
	}
}
