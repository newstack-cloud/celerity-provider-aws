//go:build unit

package utils

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type NetworkSuite struct {
	suite.Suite
}

func (s *NetworkSuite) Test_calculate_ipv4_subnet_cidr_blocks() {
	testCases := []struct {
		name         string
		vpcCIDRBlock string
		numSubnets   int
		expected     []string
	}{
		{
			name:         "6 subnets for 10.0.0.0/16",
			vpcCIDRBlock: "10.0.0.0/16",
			numSubnets:   6,
			expected: []string{
				"10.0.0.0/19",
				"10.0.32.0/19",
				"10.0.64.0/19",
				"10.0.96.0/19",
				"10.0.128.0/19",
				"10.0.160.0/19",
			},
		},
		{
			name:         "1 subnet for 10.0.0.0/24",
			vpcCIDRBlock: "10.0.0.0/24",
			numSubnets:   1,
			expected: []string{
				"10.0.0.0/24",
			},
		},
		{
			name:         "3 subnets for 10.0.0.0/26",
			vpcCIDRBlock: "10.0.0.0/26",
			numSubnets:   3,
			expected: []string{
				"10.0.0.0/28",
				"10.0.0.16/28",
				"10.0.0.32/28",
			},
		},
		{
			name:         "24 subnets for 10.0.0.0/16",
			vpcCIDRBlock: "10.0.0.0/16",
			numSubnets:   24,
			expected: []string{
				"10.0.0.0/21",
				"10.0.8.0/21",
				"10.0.16.0/21",
				"10.0.24.0/21",
				"10.0.32.0/21",
				"10.0.40.0/21",
				"10.0.48.0/21",
				"10.0.56.0/21",
				"10.0.64.0/21",
				"10.0.72.0/21",
				"10.0.80.0/21",
				"10.0.88.0/21",
				"10.0.96.0/21",
				"10.0.104.0/21",
				"10.0.112.0/21",
				"10.0.120.0/21",
				"10.0.128.0/21",
				"10.0.136.0/21",
				"10.0.144.0/21",
				"10.0.152.0/21",
				"10.0.160.0/21",
				"10.0.168.0/21",
				"10.0.176.0/21",
				"10.0.184.0/21",
			},
		},
	}

	for _, testCase := range testCases {
		s.Run(testCase.name, func() {
			subnetCIDRBlocks, err := CalculateIPv4SubnetCIDRBlocks(
				testCase.vpcCIDRBlock,
				testCase.numSubnets,
			)
			s.NoError(err)
			s.Equal(testCase.expected, subnetCIDRBlocks)
		})
	}
}

func (s *NetworkSuite) Test_calculate_ipv6_subnet_cidr_blocks() {
	testCases := []struct {
		name         string
		vpcCIDRBlock string
		numSubnets   int
		expected     []string
	}{
		{
			name:         "6 subnets for 2001:db8:1234:1a00::/56 with /64 subnet prefix",
			vpcCIDRBlock: "2001:db8:1234:1a00::/56",
			numSubnets:   6,
			expected: []string{
				"2001:db8:1234:1a00::/64",
				"2001:db8:1234:1a01::/64",
				"2001:db8:1234:1a02::/64",
				"2001:db8:1234:1a03::/64",
				"2001:db8:1234:1a04::/64",
				"2001:db8:1234:1a05::/64",
			},
		},
	}

	for _, testCase := range testCases {
		s.Run(testCase.name, func() {
			subnetCIDRBlocks, err := CalculateIPv6SubnetCIDRBlocks(
				testCase.vpcCIDRBlock,
				testCase.numSubnets,
				64,
			)
			s.NoError(err)
			s.Equal(testCase.expected, subnetCIDRBlocks)
		})
	}
}

func TestNetworkSuite(t *testing.T) {
	suite.Run(t, new(NetworkSuite))
}
