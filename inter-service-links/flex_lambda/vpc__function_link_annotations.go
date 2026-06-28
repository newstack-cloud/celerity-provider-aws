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
				"function reach private in-VPC resources (a linked database or cache) and, where the VPC " +
				"provides a NAT gateway, the public internet. \"public\" lets the function reach in-VPC " +
				"resources but gives it no outbound internet access, avoiding NAT gateway cost for VPCs " +
				"provisioned without one.",
			DefaultValue: core.ScalarFromString("private"),
			AllowedValues: []*core.ScalarValue{
				core.ScalarFromString("private"),
				core.ScalarFromString("public"),
			},
			Required: false,
		},
	}
}
