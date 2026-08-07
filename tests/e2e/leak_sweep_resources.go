//go:build integration

package e2e

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/newstack-cloud/bluelink-provider-aws/flex"
)

func (c *sweepClients) leakedLambdaFunctions(
	ctx context.Context,
	scopes []string,
) []leakedResource {
	leaks := []leakedResource{}
	paginator := lambda.NewListFunctionsPaginator(c.lambda, &lambda.ListFunctionsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			sweepWarn("list lambda functions", err)
			return leaks
		}
		for _, function := range page.Functions {
			name := aws.ToString(function.FunctionName)
			if nameInScopes(name, scopes) {
				leaks = append(leaks, newLeak(kindLambdaFunction, name))
			}
		}
	}

	return leaks
}

func (c *sweepClients) leakedIAMRoles(ctx context.Context, scopes []string) []leakedResource {
	leaks := []leakedResource{}
	paginator := iam.NewListRolesPaginator(c.iam, &iam.ListRolesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			sweepWarn("list iam roles", err)
			return leaks
		}
		for _, role := range page.Roles {
			name := aws.ToString(role.RoleName)
			if nameInScopes(name, scopes) {
				leaks = append(leaks, newLeak(kindIAMRole, name))
			}
		}
	}

	return leaks
}

func (c *sweepClients) leakedQueues(ctx context.Context, scopes []string) []leakedResource {
	leaks := []leakedResource{}
	paginator := sqs.NewListQueuesPaginator(c.sqs, &sqs.ListQueuesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			sweepWarn("list sqs queues", err)
			return leaks
		}
		for _, url := range page.QueueUrls {
			name := url[strings.LastIndex(url, "/")+1:]
			if nameInScopes(name, scopes) {
				leaks = append(leaks, leakedResource{
					kind:   kindSQSQueue,
					id:     name,
					handle: url,
				})
			}
		}
	}

	return leaks
}

func (c *sweepClients) leakedTables(ctx context.Context, scopes []string) []leakedResource {
	leaks := []leakedResource{}
	paginator := dynamodb.NewListTablesPaginator(c.dynamodb, &dynamodb.ListTablesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			sweepWarn("list dynamodb tables", err)
			return leaks
		}
		for _, name := range page.TableNames {
			if nameInScopes(name, scopes) {
				leaks = append(leaks, newLeak(kindDynamoDBTable, name))
			}
		}
	}

	return leaks
}

func (c *sweepClients) leakedRules(ctx context.Context, scopes []string) []leakedResource {
	output, err := c.eventbridge.ListRules(ctx, &eventbridge.ListRulesInput{})
	if err != nil {
		sweepWarn("list eventbridge rules", err)
		return nil
	}

	leaks := []leakedResource{}
	for _, rule := range output.Rules {
		name := aws.ToString(rule.Name)
		if nameInScopes(name, scopes) {
			leaks = append(leaks, newLeak(kindEventBridgeRule, name))
		}
	}

	return leaks
}

func (c *sweepClients) leakedRDSClusters(ctx context.Context, scopes []string) []leakedResource {
	return c.leakedCloudControlResources(ctx, scopes, "AWS::RDS::DBCluster", kindRDSCluster)
}

func (c *sweepClients) leakedRDSSubnetGroups(
	ctx context.Context,
	scopes []string,
) []leakedResource {
	return c.leakedCloudControlResources(
		ctx,
		scopes,
		"AWS::RDS::DBSubnetGroup",
		kindRDSSubnetGroup,
	)
}

// RDS resources are managed through Cloud Control rather than a dedicated client, so
// they are listed the same way. The identifier Cloud Control returns is the resource
// name for both of these types, which is what the scope prefix is matched against.
func (c *sweepClients) leakedCloudControlResources(
	ctx context.Context,
	scopes []string,
	typeName string,
	kind string,
) []leakedResource {
	output, err := c.cloudcontrol.ListResources(ctx, &cloudcontrol.ListResourcesInput{
		TypeName: aws.String(typeName),
	})
	if err != nil {
		sweepWarn("list "+typeName, err)
		return nil
	}

	leaks := []leakedResource{}
	for _, descriptor := range output.ResourceDescriptions {
		identifier := aws.ToString(descriptor.Identifier)
		if nameInScopes(identifier, scopes) {
			leaks = append(leaks, newLeak(kind, identifier))
		}
	}

	return leaks
}

// Flex VPCs are matched on the name tag the provider stamps rather than on a resource
// name, since a VPC has no name of its own.
func (c *sweepClients) leakedFlexVPCs(ctx context.Context, scopes []string) []leakedResource {
	output, err := c.ec2.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("tag-key"),
				Values: []string{flex.TagFlexVPCName},
			},
		},
	})
	if err != nil {
		sweepWarn("describe vpcs", err)
		return nil
	}

	leaks := []leakedResource{}
	for _, vpc := range output.Vpcs {
		name, hasName := TagValue(vpc.Tags, flex.TagFlexVPCName)
		if hasName && nameInScopes(name, scopes) {
			leaks = append(leaks, leakedResource{
				kind:   kindFlexVPC,
				id:     name,
				handle: aws.ToString(vpc.VpcId),
			})
		}
	}

	return leaks
}

func (c *sweepClients) deleteFunction(ctx context.Context, name string) error {
	_, err := c.lambda.DeleteFunction(ctx, &lambda.DeleteFunctionInput{
		FunctionName: aws.String(name),
	})

	return err
}

// A role cannot be deleted while policies are still attached to it, and the links
// under test attach both managed and inline policies.
func (c *sweepClients) deleteRole(ctx context.Context, name string) error {
	attached, err := c.iam.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(name),
	})
	if err == nil {
		for _, policy := range attached.AttachedPolicies {
			_, err = c.iam.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{
				RoleName:  aws.String(name),
				PolicyArn: policy.PolicyArn,
			})
			if err != nil {
				return err
			}
		}
	}

	inline, err := c.iam.ListRolePolicies(ctx, &iam.ListRolePoliciesInput{
		RoleName: aws.String(name),
	})
	if err == nil {
		for _, policyName := range inline.PolicyNames {
			_, err = c.iam.DeleteRolePolicy(ctx, &iam.DeleteRolePolicyInput{
				RoleName:   aws.String(name),
				PolicyName: aws.String(policyName),
			})
			if err != nil {
				return err
			}
		}
	}

	_, err = c.iam.DeleteRole(ctx, &iam.DeleteRoleInput{RoleName: aws.String(name)})

	return err
}

func (c *sweepClients) deleteQueue(ctx context.Context, queueURL string) error {
	_, err := c.sqs.DeleteQueue(ctx, &sqs.DeleteQueueInput{QueueUrl: aws.String(queueURL)})

	return err
}

func (c *sweepClients) deleteTable(ctx context.Context, name string) error {
	_, err := c.dynamodb.DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String(name),
	})

	return err
}

// A rule holding targets cannot be deleted, and every rule under test has at least one.
func (c *sweepClients) deleteRule(ctx context.Context, name string) error {
	targets, err := c.eventbridge.ListTargetsByRule(ctx, &eventbridge.ListTargetsByRuleInput{
		Rule: aws.String(name),
	})
	if err == nil && len(targets.Targets) > 0 {
		ids := make([]string, 0, len(targets.Targets))
		for _, target := range targets.Targets {
			ids = append(ids, aws.ToString(target.Id))
		}
		_, err = c.eventbridge.RemoveTargets(ctx, &eventbridge.RemoveTargetsInput{
			Rule: aws.String(name),
			Ids:  ids,
		})
		if err != nil {
			return err
		}
	}

	_, err = c.eventbridge.DeleteRule(ctx, &eventbridge.DeleteRuleInput{
		Name: aws.String(name),
	})

	return err
}

func (c *sweepClients) deleteCloudControlResource(
	ctx context.Context,
	typeName string,
	identifier string,
) error {
	_, err := c.cloudcontrol.DeleteResource(ctx, &cloudcontrol.DeleteResourceInput{
		TypeName:   aws.String(typeName),
		Identifier: aws.String(identifier),
	})

	return err
}
