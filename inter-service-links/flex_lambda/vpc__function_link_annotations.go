package flexlambda

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func vpcFunctionLinkAnnotations() map[string]*provider.LinkAnnotationDefinition {
	return map[string]*provider.LinkAnnotationDefinition{
		// Placement config is per-function, so the annotation is on resource B (the
		// function being placed). A VPC can place many functions, each in a tier.
		"aws/lambda/function::aws.flexvpc.lambda.subnetType": {
			Name:  "aws.flexvpc.lambda.subnetType",
			Label: "Subnet Type",
			Type:  core.ScalarTypeString,
			Description: "Which subnet tier to place the function in. \"private\" (default) lets the " +
				"function reach private in-VPC resources it is linked to and, where the VPC provides " +
				"a NAT gateway, the internet over both IPv4 and IPv6. \"public\" lets the function " +
				"reach in-VPC resources it is linked to and the internet over IPv6 only: a " +
				"VPC-attached function is never assigned a public IPv4 address, so it cannot use the " +
				"internet gateway for IPv4, but its IPv6 address is globally routable. \"public\" " +
				"therefore avoids NAT gateway cost at the price of IPv4 reachability.",
			DefaultValue: core.ScalarFromString("private"),
			AllowedValues: []*core.ScalarValue{
				core.ScalarFromString("private"),
				core.ScalarFromString("public"),
			},
			Required: false,
		},
		"aws/lambda/function::aws.flexvpc.lambda.egress": {
			Name:  "aws.flexvpc.lambda.egress",
			Label: "Outbound Access",
			Type:  core.ScalarTypeString,
			Description: "What the function may reach outside the VPC. \"internet\" opens all " +
				"destinations the VPC can route to, \"none\" opens none, and a comma-separated list " +
				"of CIDR ranges opens only those. Reach to resources inside the VPC is not " +
				"controlled here; it is granted by the links between the function and those " +
				"resources.\n\nLeaving this unset is not the same as \"none\": unset keeps whatever " +
				"the VPC's topology can deliver, so placing a function in a VPC never silently " +
				"removes outbound access it had before. Set it only to narrow.\n\nRequesting egress " +
				"a VPC cannot provide, such as \"internet\" on a VPC with no outbound path, is an " +
				"error rather than a rule that grants nothing.",
			Required: false,
		},
	}
}
