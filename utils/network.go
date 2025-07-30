package utils

import (
	"errors"
	"math/big"
	"math/bits"
	"net"
)

// CalculateIPv4SubnetCIDRBlocks calculates the IPv4 CIDR blocks for a given number of subnets
// to be deployed to the VPC with the given CIDR block.
func CalculateIPv4SubnetCIDRBlocks(vpcCIDRBlock string, numSubnets int) ([]string, error) {
	if numSubnets == 1 {
		return []string{vpcCIDRBlock}, nil
	}

	subnetBits := bits.Len(uint(numSubnets))
	_, vpcNet, err := net.ParseCIDR(vpcCIDRBlock)
	if err != nil {
		return []string{}, err
	}

	ones, _ := vpcNet.Mask.Size()
	subnetPrefix := ones + subnetBits
	subnetMask := net.CIDRMask(ones+subnetBits, 32)

	baseIP := vpcNet.IP.To4()
	if baseIP == nil {
		return []string{}, errors.New("only Ipv4 is supported for VPC CIDR blocks")
	}

	baseIPInt := big.NewInt(0).SetBytes(baseIP)
	subnetCIDRBlocks := []string{}
	for i := range numSubnets {
		subnetSize := 1 << (32 - subnetPrefix)
		subnetIPInt := big.NewInt(0).Add(baseIPInt, big.NewInt(int64(i*subnetSize)))
		subnetIP := net.IP(subnetIPInt.Bytes())
		// Ensure 4 bytes for an IPv4 address.
		if len(subnetIP) < 4 {
			pad := make([]byte, 4-len(subnetIP))
			subnetIP = append(subnetIP, pad...)
		}
		subnetCIDR := net.IPNet{
			IP:   subnetIP,
			Mask: subnetMask,
		}
		subnetCIDRBlocks = append(subnetCIDRBlocks, subnetCIDR.String())
	}

	return subnetCIDRBlocks, nil
}

// CalculateIPv6SubnetCIDRBlocks calculates the IPv6 CIDR blocks for a given number of subnets
// to be deployed to the VPC with the given CIDR block.
func CalculateIPv6SubnetCIDRBlocks(
	vpcCIDRBlock string,
	numSubnets int,
	subnetPrefix int,
) ([]string, error) {
	_, vpcNet, err := net.ParseCIDR(vpcCIDRBlock)
	if err != nil {
		return []string{}, err
	}
	if vpcNet.IP.To16() == nil || vpcNet.IP.To4() != nil {
		return nil, errors.New("the provided CIDR block is not a valid IPv6 CIDR block")
	}

	subnetMask := net.CIDRMask(subnetPrefix, 128)

	baseIPInt := big.NewInt(0).SetBytes(vpcNet.IP)
	prefixLen, _ := vpcNet.Mask.Size()
	subnetBits := subnetPrefix - prefixLen
	if subnetBits < 0 {
		return nil, errors.New("the provided subnet prefix must be larger than the base prefix")
	}
	if numSubnets > (1 << subnetBits) {
		return nil, errors.New("the provided number of subnets is too large for the given prefix")
	}

	subnetSize := big.NewInt(0).Lsh(big.NewInt(1), uint(128-subnetPrefix))
	ipv6SubnetCIDRBlocks := []string{}
	for i := range numSubnets {
		subnetIP := big.NewInt(0).Add(baseIPInt, big.NewInt(0).Mul(subnetSize, big.NewInt(int64(i))))
		ip := subnetIP.Bytes()
		// Ensure 16 bytes for an IPv6 address.
		if len(ip) < 16 {
			pad := make([]byte, 16-len(ip))
			ip = append(ip, pad...)
		}
		subnetCIDR := net.IPNet{
			IP:   net.IP(ip),
			Mask: subnetMask,
		}
		ipv6SubnetCIDRBlocks = append(ipv6SubnetCIDRBlocks, subnetCIDR.String())
	}

	return ipv6SubnetCIDRBlocks, nil
}
