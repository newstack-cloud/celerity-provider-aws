//go:build unit

package flexlambda

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

// Lambda validates the execution role's network interface permissions at the moment the
// VPC attachment is set, so the grant is only useful if it lands first. These cover the
// ordering in both directions, since getting either one backwards leaves a function that
// cannot be attached, or a VPC that cannot be deleted.
type VPCFunctionENIGrantSuite struct {
	suite.Suite
}

// Built through the exported link constructor rather than by assembling its actions,
// so the tests go through the same path the provider registers.
func (s *VPCFunctionENIGrantSuite) linkActions(
	iamSvc iamservice.Service,
	lambdaSvc lambdaservice.Service,
) provider.Link {
	loader := &testutils.MockAWSConfigLoader{}
	build := vpcFunctionLinkFactory(iamSvc)

	return build(pluginutils.LinkServiceDeps[
		*aws.Config, ec2service.Service, *aws.Config, lambdaservice.Service,
	]{
		ResourceAService: pluginutils.ServiceWithConfigStore[*aws.Config, ec2service.Service]{
			ServiceFactory: placementEC2ServiceFactory(),
			ConfigStore:    testConfigStore(loader),
		},
		ResourceBService: pluginutils.ServiceWithConfigStore[*aws.Config, lambdaservice.Service]{
			ServiceFactory: func(c *aws.Config, pc provider.Context) lambdaservice.Service {
				return lambdaSvc
			},
			ConfigStore: testConfigStore(loader),
		},
	})
}

func (s *VPCFunctionENIGrantSuite) placeInput() *provider.LinkUpdateResourceInput {
	return &provider.LinkUpdateResourceInput{
		LinkUpdateType:    provider.LinkUpdateTypeCreate,
		ResourceInfo:      functionResourceInfoB(),
		OtherResourceInfo: flexVPCResourceInfoA(),
		LinkContext:       testLinkContext(),
		LinkID:            vfLinkID,
		ResourceService:   vfRoleService(),
	}
}

// A role whose Bluelink inline policy already carries this link's ENI statement, which is
// what a destroy runs against.
func vfGrantedRoleIamMock() *iammock.Mock {
	// Spelled out rather than built with the provider's own statement helper: a
	// fixture that changes shape whenever the implementation does cannot show that a
	// destroy removed the statement the implementation had granted.
	document, _ := json.Marshal(map[string]any{
		"Version": "2012-10-17",
		"Statement": []map[string]any{
			{
				"Sid":    vfENIStatement,
				"Effect": "Allow",
				"Action": []string{
					"ec2:CreateNetworkInterface",
					"ec2:DescribeNetworkInterfaces",
					"ec2:DescribeSubnets",
					"ec2:DeleteNetworkInterface",
					"ec2:AssignPrivateIpAddresses",
					"ec2:UnassignPrivateIpAddresses",
				},
				"Resource": "*",
			},
		},
	})

	return iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{
			PolicyNames: []string{linkutils.InlineAccessPolicyName()},
		}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithGetRolePolicyOutput(&iam.GetRolePolicyOutput{
			RoleName:       aws.String(vfRoleName),
			PolicyName:     aws.String(linkutils.InlineAccessPolicyName()),
			PolicyDocument: aws.String(string(document)),
		}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
		iammock.WithDeleteRolePolicyOutput(&iam.DeleteRolePolicyOutput{}),
	)
}

// If the attachment were set first, Lambda would reject it for a permission the link is
// about to grant, and the link would fail on a condition it created itself.
func (s *VPCFunctionENIGrantSuite) Test_does_not_attach_the_function_when_the_grant_fails() {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(vfGetFunctionOutput()),
	)
	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyError(errors.New("access denied writing the role policy")),
	)

	actions := s.linkActions(iamSvc, lambdaSvc)
	_, err := actions.UpdateResourceB(context.Background(), s.placeInput())

	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "network interface permissions")
	lambdaSvc.AssertNotCalled(&s.Suite, "UpdateFunctionConfiguration")
}

// The granted statement has to be the one Lambda actually checks for. A policy that omits
// any of these actions is accepted by IAM and then rejected by Lambda at attach time.
func (s *VPCFunctionENIGrantSuite) Test_grants_the_network_interface_actions_lambda_requires() {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(vfGetFunctionOutput()),
		lambdamock.WithUpdateFunctionConfigurationOutput(nil),
	)
	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)

	actions := s.linkActions(iamSvc, lambdaSvc)
	_, err := actions.UpdateResourceB(context.Background(), s.placeInput())
	s.Require().NoError(err)

	iamSvc.AssertCalledWith(
		&s.Suite,
		"PutRolePolicy",
		0,
		plugintestutils.Any,
		matchENIGrantStatement,
	)
}

// Lambda deletes the function's network interfaces using the execution role, so revoking
// before detaching strands them. A stranded interface holds the VPC's security groups,
// and with them the VPC, undeletable.
func (s *VPCFunctionENIGrantSuite) Test_does_not_revoke_the_grant_when_the_detach_fails() {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(vfGetFunctionOutput()),
		lambdamock.WithUpdateFunctionConfigurationError(
			errors.New("the function could not be detached"),
		),
	)
	iamSvc := vfGrantedRoleIamMock()

	input := s.placeInput()
	input.LinkUpdateType = provider.LinkUpdateTypeDestroy

	actions := s.linkActions(iamSvc, lambdaSvc)
	_, err := actions.UpdateResourceB(context.Background(), input)

	// The permissions have to outlive an attachment that is still in place.
	s.Require().Error(err)
	iamSvc.AssertNotCalled(&s.Suite, "PutRolePolicy")
	iamSvc.AssertNotCalled(&s.Suite, "DeleteRolePolicy")
}

func (s *VPCFunctionENIGrantSuite) Test_revokes_the_grant_once_the_function_is_detached() {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(vfGetFunctionOutput()),
		lambdamock.WithUpdateFunctionConfigurationOutput(nil),
	)
	iamSvc := vfGrantedRoleIamMock()

	input := s.placeInput()
	input.LinkUpdateType = provider.LinkUpdateTypeDestroy

	actions := s.linkActions(iamSvc, lambdaSvc)
	_, err := actions.UpdateResourceB(context.Background(), input)
	s.Require().NoError(err)

	// The statement was the policy's only one, so removing it removes the policy.
	iamSvc.AssertCalled(&s.Suite, "DeleteRolePolicy")
}

// A function whose execution role is not a blueprint resource is left alone: the link has
// no role it is entitled to modify, and the user's own role carries the permissions.
func (s *VPCFunctionENIGrantSuite) Test_places_the_function_without_a_grant_when_there_is_no_role() {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(vfGetFunctionOutput()),
		lambdamock.WithUpdateFunctionConfigurationOutput(nil),
	)
	iamSvc := iammock.CreateIamServiceMock()

	input := s.placeInput()
	// No role resource in state, which is what an externally-managed role looks like.
	input.ResourceService = resourceservicemock.Create()

	actions := s.linkActions(iamSvc, lambdaSvc)
	output, err := actions.UpdateResourceB(context.Background(), input)
	s.Require().NoError(err)

	iamSvc.AssertNotCalled(&s.Suite, "PutRolePolicy")
	s.Assert().NotContains(output.LinkData.Fields, vfExecRoleField)
	lambdaSvc.AssertCalled(&s.Suite, "UpdateFunctionConfiguration")
}

func matchENIGrantStatement(arg any) bool {
	input, ok := arg.(*iam.PutRolePolicyInput)
	if !ok {
		return false
	}
	if aws.ToString(input.RoleName) != vfRoleName ||
		aws.ToString(input.PolicyName) != linkutils.InlineAccessPolicyName() {
		return false
	}

	var doc struct {
		Statement []struct {
			Sid      string
			Effect   string
			Action   []string
			Resource string
		}
	}
	if err := json.Unmarshal([]byte(aws.ToString(input.PolicyDocument)), &doc); err != nil {
		return false
	}

	for _, statement := range doc.Statement {
		if statement.Sid != vfENIStatement {
			continue
		}
		return statement.Effect == "Allow" &&
			statement.Resource == "*" &&
			hasAllActions(statement.Action, []string{
				"ec2:CreateNetworkInterface",
				"ec2:DescribeNetworkInterfaces",
				"ec2:DescribeSubnets",
				"ec2:DeleteNetworkInterface",
				"ec2:AssignPrivateIpAddresses",
				"ec2:UnassignPrivateIpAddresses",
			})
	}

	return false
}

func hasAllActions(have, want []string) bool {
	set := map[string]bool{}
	for _, action := range have {
		set[action] = true
	}
	for _, action := range want {
		if !set[action] {
			return false
		}
	}
	return true
}

func TestVPCFunctionENIGrantSuite(t *testing.T) {
	suite.Run(t, new(VPCFunctionENIGrantSuite))
}
