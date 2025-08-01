package ec2mock

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
)

type ec2ServiceMock struct {
	plugintestutils.MockCalls

	// VPC-related mock fields
	createVpcOutput     *ec2.CreateVpcOutput
	createVpcError      error
	describeVpcsOutputs []*ec2.DescribeVpcsOutput
	describeVpcsError   error
	deleteVpcOutput     *ec2.DeleteVpcOutput
	deleteVpcError      error

	// VPC attribute-related mock fields
	modifyVpcAttributeOutput   *ec2.ModifyVpcAttributeOutput
	modifyVpcAttributeError    error
	describeVpcAttributeOutput *ec2.DescribeVpcAttributeOutput
	describeVpcAttributeError  error

	// Subnet-related mock fields
	createSubnetOutputs         []*ec2.CreateSubnetOutput
	createSubnetError           error
	describeSubnetsOutput       *ec2.DescribeSubnetsOutput
	describeSubnetsError        error
	deleteSubnetOutput          *ec2.DeleteSubnetOutput
	deleteSubnetError           error
	modifySubnetAttributeOutput *ec2.ModifySubnetAttributeOutput
	modifySubnetAttributeError  error

	// Availability zone-related mock fields
	describeAvailabilityZonesOutput *ec2.DescribeAvailabilityZonesOutput
	describeAvailabilityZonesError  error

	// Route table-related mock fields
	createRouteTableOutputs      []*ec2.CreateRouteTableOutput
	createRouteTableError        error
	describeRouteTablesOutput    *ec2.DescribeRouteTablesOutput
	describeRouteTablesError     error
	deleteRouteTableOutput       *ec2.DeleteRouteTableOutput
	deleteRouteTableError        error
	associateRouteTableOutput    *ec2.AssociateRouteTableOutput
	associateRouteTableError     error
	disassociateRouteTableOutput *ec2.DisassociateRouteTableOutput
	disassociateRouteTableError  error
	createRouteOutput            *ec2.CreateRouteOutput
	createRouteError             error

	// Internet gateway-related mock fields
	createInternetGatewayOutput    *ec2.CreateInternetGatewayOutput
	createInternetGatewayError     error
	describeInternetGatewaysOutput *ec2.DescribeInternetGatewaysOutput
	describeInternetGatewaysError  error
	deleteInternetGatewayOutput    *ec2.DeleteInternetGatewayOutput
	deleteInternetGatewayError     error
	attachInternetGatewayOutput    *ec2.AttachInternetGatewayOutput
	attachInternetGatewayError     error
	detachInternetGatewayOutput    *ec2.DetachInternetGatewayOutput
	detachInternetGatewayError     error

	// NAT gateway-related mock fields
	createNatGatewayOutputs   []*ec2.CreateNatGatewayOutput
	createNatGatewayError     error
	describeNatGatewaysOutput *ec2.DescribeNatGatewaysOutput
	describeNatGatewaysError  error
	deleteNatGatewayOutput    *ec2.DeleteNatGatewayOutput
	deleteNatGatewayError     error

	// Elastic IP-related mock fields
	allocateAddressOutputs []*ec2.AllocateAddressOutput
	allocateAddressError   error
	releaseAddressOutput   *ec2.ReleaseAddressOutput
	releaseAddressError    error

	// Security group-related mock fields
	createSecurityGroupOutput       *ec2.CreateSecurityGroupOutput
	createSecurityGroupError        error
	describeSecurityGroupsOutput    *ec2.DescribeSecurityGroupsOutput
	describeSecurityGroupsError     error
	deleteSecurityGroupOutput       *ec2.DeleteSecurityGroupOutput
	deleteSecurityGroupError        error
	revokeSecurityGroupEgressOutput *ec2.RevokeSecurityGroupEgressOutput
	revokeSecurityGroupEgressError  error

	// Network ACL-related mock fields
	createNetworkAclOutput             *ec2.CreateNetworkAclOutput
	createNetworkAclError              error
	describeNetworkAclsOutput          *ec2.DescribeNetworkAclsOutput
	describeNetworkAclsError           error
	deleteNetworkAclOutput             *ec2.DeleteNetworkAclOutput
	deleteNetworkAclError              error
	createNetworkAclEntryOutput        *ec2.CreateNetworkAclEntryOutput
	createNetworkAclEntryError         error
	replaceNetworkAclAssociationOutput *ec2.ReplaceNetworkAclAssociationOutput
	replaceNetworkAclAssociationError  error

	// VPC endpoint-related mock fields
	describeVpcEndpointsOutput *ec2.DescribeVpcEndpointsOutput
	describeVpcEndpointsError  error
	createVpcEndpointOutput    *ec2.CreateVpcEndpointOutput
	createVpcEndpointError     error
	modifyVpcEndpointOutput    *ec2.ModifyVpcEndpointOutput
	modifyVpcEndpointError     error
	deleteVpcEndpointsOutput   *ec2.DeleteVpcEndpointsOutput
	deleteVpcEndpointsError    error

	// Security group ingress-related mock fields
	authorizeSecurityGroupIngressOutput *ec2.AuthorizeSecurityGroupIngressOutput
	authorizeSecurityGroupIngressError  error

	// Tag-related mock fields
	createTagsOutput *ec2.CreateTagsOutput
	createTagsError  error
	deleteTagsOutput *ec2.DeleteTagsOutput
	deleteTagsError  error

	// Tracking fields for sequenced stubs.
	describeVpcsCallCount     int
	createSubnetCallCount     int
	createRouteTableCallCount int
	allocateAddressCallCount  int
	createNatGatewayCallCount int

	mu sync.Mutex
}

type ec2ServiceMockOption func(*ec2ServiceMock)

func CreateEc2ServiceMockFactory(
	opts ...ec2ServiceMockOption,
) func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
	return func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
		return CreateEc2ServiceMock(opts...)
	}
}

func CreateEc2ServiceMock(
	opts ...ec2ServiceMockOption,
) *ec2ServiceMock {
	mock := &ec2ServiceMock{}
	for _, opt := range opts {
		opt(mock)
	}
	return mock
}

func NewMockService(t any) *ec2ServiceMock {
	return &ec2ServiceMock{}
}

// VPC-related options.
func WithCreateVpcOutput(output *ec2.CreateVpcOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.createVpcOutput = output
	}
}

func WithCreateVpcError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.createVpcError = err
	}
}

func WithDescribeVpcsOutputs(outputs []*ec2.DescribeVpcsOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.describeVpcsOutputs = outputs
	}
}

func WithDescribeVpcsError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.describeVpcsError = err
	}
}

func WithDeleteVpcOutput(output *ec2.DeleteVpcOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.deleteVpcOutput = output
	}
}

func WithDeleteVpcError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.deleteVpcError = err
	}
}

// VPC attribute-related options.
func WithModifyVpcAttributeOutput(output *ec2.ModifyVpcAttributeOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.modifyVpcAttributeOutput = output
	}
}

func WithModifyVpcAttributeError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.modifyVpcAttributeError = err
	}
}

func WithDescribeVpcAttributeOutput(output *ec2.DescribeVpcAttributeOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.describeVpcAttributeOutput = output
	}
}

func WithDescribeVpcAttributeError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.describeVpcAttributeError = err
	}
}

// Subnet-related options.
func WithCreateSubnetOutputs(outputs []*ec2.CreateSubnetOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.createSubnetOutputs = outputs
	}
}

func WithCreateSubnetError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.createSubnetError = err
	}
}

func WithDescribeSubnetsOutput(output *ec2.DescribeSubnetsOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.describeSubnetsOutput = output
	}
}

func WithDescribeSubnetsError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.describeSubnetsError = err
	}
}

func WithDeleteSubnetOutput(output *ec2.DeleteSubnetOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.deleteSubnetOutput = output
	}
}

func WithDeleteSubnetError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.deleteSubnetError = err
	}
}

func WithModifySubnetAttributeOutput(output *ec2.ModifySubnetAttributeOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.modifySubnetAttributeOutput = output
	}
}

func WithModifySubnetAttributeError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.modifySubnetAttributeError = err
	}
}

// Availability zone-related options.
func WithDescribeAvailabilityZonesOutput(output *ec2.DescribeAvailabilityZonesOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.describeAvailabilityZonesOutput = output
	}
}

func WithDescribeAvailabilityZonesError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.describeAvailabilityZonesError = err
	}
}

// Route table-related options.
func WithCreateRouteTableOutputs(outputs []*ec2.CreateRouteTableOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.createRouteTableOutputs = outputs
	}
}

func WithCreateRouteTableError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.createRouteTableError = err
	}
}

func WithDescribeRouteTablesOutput(output *ec2.DescribeRouteTablesOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.describeRouteTablesOutput = output
	}
}

func WithDescribeRouteTablesError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.describeRouteTablesError = err
	}
}

func WithDeleteRouteTableOutput(output *ec2.DeleteRouteTableOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.deleteRouteTableOutput = output
	}
}

func WithDeleteRouteTableError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.deleteRouteTableError = err
	}
}

func WithAssociateRouteTableOutput(output *ec2.AssociateRouteTableOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.associateRouteTableOutput = output
	}
}

func WithAssociateRouteTableError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.associateRouteTableError = err
	}
}

func WithDisassociateRouteTableOutput(output *ec2.DisassociateRouteTableOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.disassociateRouteTableOutput = output
	}
}

func WithDisassociateRouteTableError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.disassociateRouteTableError = err
	}
}

func WithCreateRouteOutput(output *ec2.CreateRouteOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.createRouteOutput = output
	}
}

func WithCreateRouteError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.createRouteError = err
	}
}

// Internet gateway-related options.
func WithCreateInternetGatewayOutput(output *ec2.CreateInternetGatewayOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.createInternetGatewayOutput = output
	}
}

func WithCreateInternetGatewayError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.createInternetGatewayError = err
	}
}

func WithDescribeInternetGatewaysOutput(output *ec2.DescribeInternetGatewaysOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.describeInternetGatewaysOutput = output
	}
}

func WithDescribeInternetGatewaysError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.describeInternetGatewaysError = err
	}
}

func WithDeleteInternetGatewayOutput(output *ec2.DeleteInternetGatewayOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.deleteInternetGatewayOutput = output
	}
}

func WithDeleteInternetGatewayError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.deleteInternetGatewayError = err
	}
}

func WithAttachInternetGatewayOutput(output *ec2.AttachInternetGatewayOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.attachInternetGatewayOutput = output
	}
}

func WithAttachInternetGatewayError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.attachInternetGatewayError = err
	}
}

func WithDetachInternetGatewayOutput(output *ec2.DetachInternetGatewayOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.detachInternetGatewayOutput = output
	}
}

func WithDetachInternetGatewayError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.detachInternetGatewayError = err
	}
}

// NAT gateway-related options.
func WithCreateNatGatewayOutputs(outputs []*ec2.CreateNatGatewayOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.createNatGatewayOutputs = outputs
	}
}

func WithCreateNatGatewayError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.createNatGatewayError = err
	}
}

func WithDescribeNatGatewaysOutput(output *ec2.DescribeNatGatewaysOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.describeNatGatewaysOutput = output
	}
}

func WithDescribeNatGatewaysError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.describeNatGatewaysError = err
	}
}

func WithDeleteNatGatewayOutput(output *ec2.DeleteNatGatewayOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.deleteNatGatewayOutput = output
	}
}

func WithDeleteNatGatewayError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.deleteNatGatewayError = err
	}
}

// Elastic IP-related options.
func WithAllocateAddressOutputs(outputs []*ec2.AllocateAddressOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.allocateAddressOutputs = outputs
	}
}

func WithAllocateAddressError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.allocateAddressError = err
	}
}

func WithReleaseAddressOutput(output *ec2.ReleaseAddressOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.releaseAddressOutput = output
	}
}

func WithReleaseAddressError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.releaseAddressError = err
	}
}

// Security group-related options.
func WithCreateSecurityGroupOutput(output *ec2.CreateSecurityGroupOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.createSecurityGroupOutput = output
	}
}

func WithCreateSecurityGroupError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.createSecurityGroupError = err
	}
}

func WithDescribeSecurityGroupsOutput(output *ec2.DescribeSecurityGroupsOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.describeSecurityGroupsOutput = output
	}
}

func WithDescribeSecurityGroupsError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.describeSecurityGroupsError = err
	}
}

func WithDeleteSecurityGroupOutput(output *ec2.DeleteSecurityGroupOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.deleteSecurityGroupOutput = output
	}
}

func WithDeleteSecurityGroupError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.deleteSecurityGroupError = err
	}
}

func WithRevokeSecurityGroupEgressOutput(output *ec2.RevokeSecurityGroupEgressOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.revokeSecurityGroupEgressOutput = output
	}
}

func WithRevokeSecurityGroupEgressError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.revokeSecurityGroupEgressError = err
	}
}

// Network ACL-related options.
func WithCreateNetworkAclOutput(output *ec2.CreateNetworkAclOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.createNetworkAclOutput = output
	}
}

func WithCreateNetworkAclError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.createNetworkAclError = err
	}
}

func WithDescribeNetworkAclsOutput(output *ec2.DescribeNetworkAclsOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.describeNetworkAclsOutput = output
	}
}

func WithDescribeNetworkAclsError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.describeNetworkAclsError = err
	}
}

func WithDeleteNetworkAclOutput(output *ec2.DeleteNetworkAclOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.deleteNetworkAclOutput = output
	}
}

func WithDeleteNetworkAclError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.deleteNetworkAclError = err
	}
}

func WithCreateNetworkAclEntryOutput(output *ec2.CreateNetworkAclEntryOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.createNetworkAclEntryOutput = output
	}
}

func WithCreateNetworkAclEntryError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.createNetworkAclEntryError = err
	}
}

func WithReplaceNetworkAclAssociationOutput(output *ec2.ReplaceNetworkAclAssociationOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.replaceNetworkAclAssociationOutput = output
	}
}

func WithReplaceNetworkAclAssociationError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.replaceNetworkAclAssociationError = err
	}
}

// VPC endpoint-related options.
func WithDescribeVpcEndpointsOutput(output *ec2.DescribeVpcEndpointsOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.describeVpcEndpointsOutput = output
	}
}

func WithDescribeVpcEndpointsError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.describeVpcEndpointsError = err
	}
}

func WithCreateVpcEndpointOutput(output *ec2.CreateVpcEndpointOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.createVpcEndpointOutput = output
	}
}

func WithCreateVpcEndpointError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.createVpcEndpointError = err
	}
}

func WithModifyVpcEndpointOutput(output *ec2.ModifyVpcEndpointOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.modifyVpcEndpointOutput = output
	}
}

func WithModifyVpcEndpointError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.modifyVpcEndpointError = err
	}
}

func WithDeleteVpcEndpointsOutput(output *ec2.DeleteVpcEndpointsOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.deleteVpcEndpointsOutput = output
	}
}

func WithDeleteVpcEndpointsError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.deleteVpcEndpointsError = err
	}
}

// Security group ingress-related options.
func WithAuthorizeSecurityGroupIngressOutput(output *ec2.AuthorizeSecurityGroupIngressOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.authorizeSecurityGroupIngressOutput = output
	}
}

func WithAuthorizeSecurityGroupIngressError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.authorizeSecurityGroupIngressError = err
	}
}

// Tag-related options.
func WithCreateTagsOutput(output *ec2.CreateTagsOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.createTagsOutput = output
	}
}

func WithCreateTagsError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.createTagsError = err
	}
}

func WithDeleteTagsOutput(output *ec2.DeleteTagsOutput) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.deleteTagsOutput = output
	}
}

func WithDeleteTagsError(err error) ec2ServiceMockOption {
	return func(m *ec2ServiceMock) {
		m.deleteTagsError = err
	}
}

// VPC methods.
func (m *ec2ServiceMock) CreateVpc(
	ctx context.Context,
	params *ec2.CreateVpcInput,
	optFns ...func(*ec2.Options),
) (*ec2.CreateVpcOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createVpcOutput, m.createVpcError
}

func (m *ec2ServiceMock) DescribeVpcs(
	ctx context.Context,
	params *ec2.DescribeVpcsInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeVpcsOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.RegisterCall(ctx, params)
	m.describeVpcsCallCount += 1

	if len(m.describeVpcsOutputs) == 0 {
		return nil, m.describeVpcsError
	}

	index := m.describeVpcsCallCount - 1
	if index >= len(m.describeVpcsOutputs) {
		index = index % len(m.describeVpcsOutputs)
	}
	return m.describeVpcsOutputs[index], m.describeVpcsError
}

func (m *ec2ServiceMock) DeleteVpc(
	ctx context.Context,
	params *ec2.DeleteVpcInput,
	optFns ...func(*ec2.Options),
) (*ec2.DeleteVpcOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteVpcOutput, m.deleteVpcError
}

func (m *ec2ServiceMock) ModifyVpcAttribute(
	ctx context.Context,
	params *ec2.ModifyVpcAttributeInput,
	optFns ...func(*ec2.Options),
) (*ec2.ModifyVpcAttributeOutput, error) {
	m.RegisterCall(ctx, params)
	return m.modifyVpcAttributeOutput, m.modifyVpcAttributeError
}

func (m *ec2ServiceMock) DescribeVpcAttribute(
	ctx context.Context,
	params *ec2.DescribeVpcAttributeInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeVpcAttributeOutput, error) {
	m.RegisterCall(ctx, params)
	return m.describeVpcAttributeOutput, m.describeVpcAttributeError
}

// Subnet methods.
func (m *ec2ServiceMock) CreateSubnet(
	ctx context.Context,
	params *ec2.CreateSubnetInput,
	optFns ...func(*ec2.Options),
) (*ec2.CreateSubnetOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.RegisterCall(ctx, params)
	m.createSubnetCallCount += 1

	if len(m.createSubnetOutputs) == 0 {
		return nil, m.createSubnetError
	}

	index := m.createSubnetCallCount - 1
	if index >= len(m.createSubnetOutputs) {
		index = index % len(m.createSubnetOutputs)
	}
	return m.createSubnetOutputs[index], m.createSubnetError
}

func (m *ec2ServiceMock) DescribeSubnets(
	ctx context.Context,
	params *ec2.DescribeSubnetsInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeSubnetsOutput, error) {
	m.RegisterCall(ctx, params)
	return m.describeSubnetsOutput, m.describeSubnetsError
}

func (m *ec2ServiceMock) DeleteSubnet(
	ctx context.Context,
	params *ec2.DeleteSubnetInput,
	optFns ...func(*ec2.Options),
) (*ec2.DeleteSubnetOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteSubnetOutput, m.deleteSubnetError
}

func (m *ec2ServiceMock) ModifySubnetAttribute(
	ctx context.Context,
	params *ec2.ModifySubnetAttributeInput,
	optFns ...func(*ec2.Options),
) (*ec2.ModifySubnetAttributeOutput, error) {
	m.RegisterCall(ctx, params)
	return m.modifySubnetAttributeOutput, m.modifySubnetAttributeError
}

// Availability zone methods.
func (m *ec2ServiceMock) DescribeAvailabilityZones(
	ctx context.Context,
	params *ec2.DescribeAvailabilityZonesInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeAvailabilityZonesOutput, error) {
	m.RegisterCall(ctx, params)
	return m.describeAvailabilityZonesOutput, m.describeAvailabilityZonesError
}

// Route table methods.
func (m *ec2ServiceMock) CreateRouteTable(
	ctx context.Context,
	params *ec2.CreateRouteTableInput,
	optFns ...func(*ec2.Options),
) (*ec2.CreateRouteTableOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.RegisterCall(ctx, params)
	m.createRouteTableCallCount += 1

	if len(m.createRouteTableOutputs) == 0 {
		return nil, m.createRouteTableError
	}

	index := m.createRouteTableCallCount - 1
	if index >= len(m.createRouteTableOutputs) {
		index = index % len(m.createRouteTableOutputs)
	}
	return m.createRouteTableOutputs[index], m.createRouteTableError
}

func (m *ec2ServiceMock) DescribeRouteTables(
	ctx context.Context,
	params *ec2.DescribeRouteTablesInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeRouteTablesOutput, error) {
	m.RegisterCall(ctx, params)
	return m.describeRouteTablesOutput, m.describeRouteTablesError
}

func (m *ec2ServiceMock) DeleteRouteTable(
	ctx context.Context,
	params *ec2.DeleteRouteTableInput,
	optFns ...func(*ec2.Options),
) (*ec2.DeleteRouteTableOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteRouteTableOutput, m.deleteRouteTableError
}

func (m *ec2ServiceMock) AssociateRouteTable(
	ctx context.Context,
	params *ec2.AssociateRouteTableInput,
	optFns ...func(*ec2.Options),
) (*ec2.AssociateRouteTableOutput, error) {
	m.RegisterCall(ctx, params)
	return m.associateRouteTableOutput, m.associateRouteTableError
}

func (m *ec2ServiceMock) DisassociateRouteTable(
	ctx context.Context,
	params *ec2.DisassociateRouteTableInput,
	optFns ...func(*ec2.Options),
) (*ec2.DisassociateRouteTableOutput, error) {
	m.RegisterCall(ctx, params)
	return m.disassociateRouteTableOutput, m.disassociateRouteTableError
}

func (m *ec2ServiceMock) CreateRoute(
	ctx context.Context,
	params *ec2.CreateRouteInput,
	optFns ...func(*ec2.Options),
) (*ec2.CreateRouteOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createRouteOutput, m.createRouteError
}

// Internet gateway methods.
func (m *ec2ServiceMock) CreateInternetGateway(
	ctx context.Context,
	params *ec2.CreateInternetGatewayInput,
	optFns ...func(*ec2.Options),
) (*ec2.CreateInternetGatewayOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createInternetGatewayOutput, m.createInternetGatewayError
}

func (m *ec2ServiceMock) DescribeInternetGateways(
	ctx context.Context,
	params *ec2.DescribeInternetGatewaysInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeInternetGatewaysOutput, error) {
	m.RegisterCall(ctx, params)
	return m.describeInternetGatewaysOutput, m.describeInternetGatewaysError
}

func (m *ec2ServiceMock) DeleteInternetGateway(
	ctx context.Context,
	params *ec2.DeleteInternetGatewayInput,
	optFns ...func(*ec2.Options),
) (*ec2.DeleteInternetGatewayOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteInternetGatewayOutput, m.deleteInternetGatewayError
}

func (m *ec2ServiceMock) AttachInternetGateway(
	ctx context.Context,
	params *ec2.AttachInternetGatewayInput,
	optFns ...func(*ec2.Options),
) (*ec2.AttachInternetGatewayOutput, error) {
	m.RegisterCall(ctx, params)
	return m.attachInternetGatewayOutput, m.attachInternetGatewayError
}

func (m *ec2ServiceMock) DetachInternetGateway(
	ctx context.Context,
	params *ec2.DetachInternetGatewayInput,
	optFns ...func(*ec2.Options),
) (*ec2.DetachInternetGatewayOutput, error) {
	m.RegisterCall(ctx, params)
	return m.detachInternetGatewayOutput, m.detachInternetGatewayError
}

// NAT gateway methods.
func (m *ec2ServiceMock) CreateNatGateway(
	ctx context.Context,
	params *ec2.CreateNatGatewayInput,
	optFns ...func(*ec2.Options),
) (*ec2.CreateNatGatewayOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.RegisterCall(ctx, params)
	m.createNatGatewayCallCount += 1

	if len(m.createNatGatewayOutputs) == 0 {
		return nil, m.createNatGatewayError
	}

	index := m.createNatGatewayCallCount - 1
	if index >= len(m.createNatGatewayOutputs) {
		index = index % len(m.createNatGatewayOutputs)
	}
	return m.createNatGatewayOutputs[index], m.createNatGatewayError
}

func (m *ec2ServiceMock) DescribeNatGateways(
	ctx context.Context,
	params *ec2.DescribeNatGatewaysInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeNatGatewaysOutput, error) {
	m.RegisterCall(ctx, params)
	return m.describeNatGatewaysOutput, m.describeNatGatewaysError
}

func (m *ec2ServiceMock) DeleteNatGateway(
	ctx context.Context,
	params *ec2.DeleteNatGatewayInput,
	optFns ...func(*ec2.Options),
) (*ec2.DeleteNatGatewayOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteNatGatewayOutput, m.deleteNatGatewayError
}

// Elastic IP methods.
func (m *ec2ServiceMock) AllocateAddress(
	ctx context.Context,
	params *ec2.AllocateAddressInput,
	optFns ...func(*ec2.Options),
) (*ec2.AllocateAddressOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.RegisterCall(ctx, params)
	m.allocateAddressCallCount += 1

	if len(m.allocateAddressOutputs) == 0 {
		return nil, m.allocateAddressError
	}

	index := m.allocateAddressCallCount - 1
	if index >= len(m.allocateAddressOutputs) {
		index = index % len(m.allocateAddressOutputs)
	}
	return m.allocateAddressOutputs[index], m.allocateAddressError
}

func (m *ec2ServiceMock) ReleaseAddress(
	ctx context.Context,
	params *ec2.ReleaseAddressInput,
	optFns ...func(*ec2.Options),
) (*ec2.ReleaseAddressOutput, error) {
	m.RegisterCall(ctx, params)
	return m.releaseAddressOutput, m.releaseAddressError
}

// Security group methods.
func (m *ec2ServiceMock) CreateSecurityGroup(
	ctx context.Context,
	params *ec2.CreateSecurityGroupInput,
	optFns ...func(*ec2.Options),
) (*ec2.CreateSecurityGroupOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createSecurityGroupOutput, m.createSecurityGroupError
}

func (m *ec2ServiceMock) DescribeSecurityGroups(
	ctx context.Context,
	params *ec2.DescribeSecurityGroupsInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeSecurityGroupsOutput, error) {
	m.RegisterCall(ctx, params)
	return m.describeSecurityGroupsOutput, m.describeSecurityGroupsError
}

func (m *ec2ServiceMock) DeleteSecurityGroup(
	ctx context.Context,
	params *ec2.DeleteSecurityGroupInput,
	optFns ...func(*ec2.Options),
) (*ec2.DeleteSecurityGroupOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteSecurityGroupOutput, m.deleteSecurityGroupError
}

func (m *ec2ServiceMock) RevokeSecurityGroupEgress(
	ctx context.Context,
	params *ec2.RevokeSecurityGroupEgressInput,
	optFns ...func(*ec2.Options),
) (*ec2.RevokeSecurityGroupEgressOutput, error) {
	m.RegisterCall(ctx, params)
	return m.revokeSecurityGroupEgressOutput, m.revokeSecurityGroupEgressError
}

// Network ACL methods.
func (m *ec2ServiceMock) CreateNetworkAcl(
	ctx context.Context,
	params *ec2.CreateNetworkAclInput,
	optFns ...func(*ec2.Options),
) (*ec2.CreateNetworkAclOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createNetworkAclOutput, m.createNetworkAclError
}

func (m *ec2ServiceMock) DescribeNetworkAcls(
	ctx context.Context,
	params *ec2.DescribeNetworkAclsInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeNetworkAclsOutput, error) {
	m.RegisterCall(ctx, params)
	return m.describeNetworkAclsOutput, m.describeNetworkAclsError
}

func (m *ec2ServiceMock) DeleteNetworkAcl(
	ctx context.Context,
	params *ec2.DeleteNetworkAclInput,
	optFns ...func(*ec2.Options),
) (*ec2.DeleteNetworkAclOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteNetworkAclOutput, m.deleteNetworkAclError
}

func (m *ec2ServiceMock) CreateNetworkAclEntry(
	ctx context.Context,
	params *ec2.CreateNetworkAclEntryInput,
	optFns ...func(*ec2.Options),
) (*ec2.CreateNetworkAclEntryOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createNetworkAclEntryOutput, m.createNetworkAclEntryError
}

func (m *ec2ServiceMock) ReplaceNetworkAclAssociation(
	ctx context.Context,
	params *ec2.ReplaceNetworkAclAssociationInput,
	optFns ...func(*ec2.Options),
) (*ec2.ReplaceNetworkAclAssociationOutput, error) {
	m.RegisterCall(ctx, params)
	return m.replaceNetworkAclAssociationOutput, m.replaceNetworkAclAssociationError
}

// VPC endpoint methods.
func (m *ec2ServiceMock) DescribeVpcEndpoints(
	ctx context.Context,
	params *ec2.DescribeVpcEndpointsInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeVpcEndpointsOutput, error) {
	m.RegisterCall(ctx, params)
	return m.describeVpcEndpointsOutput, m.describeVpcEndpointsError
}

func (m *ec2ServiceMock) CreateVpcEndpoint(
	ctx context.Context,
	params *ec2.CreateVpcEndpointInput,
	optFns ...func(*ec2.Options),
) (*ec2.CreateVpcEndpointOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createVpcEndpointOutput, m.createVpcEndpointError
}

func (m *ec2ServiceMock) ModifyVpcEndpoint(
	ctx context.Context,
	params *ec2.ModifyVpcEndpointInput,
	optFns ...func(*ec2.Options),
) (*ec2.ModifyVpcEndpointOutput, error) {
	m.RegisterCall(ctx, params)
	return m.modifyVpcEndpointOutput, m.modifyVpcEndpointError
}

func (m *ec2ServiceMock) DeleteVpcEndpoints(
	ctx context.Context,
	params *ec2.DeleteVpcEndpointsInput,
	optFns ...func(*ec2.Options),
) (*ec2.DeleteVpcEndpointsOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteVpcEndpointsOutput, m.deleteVpcEndpointsError
}

// Security group ingress methods.
func (m *ec2ServiceMock) AuthorizeSecurityGroupIngress(
	ctx context.Context,
	params *ec2.AuthorizeSecurityGroupIngressInput,
	optFns ...func(*ec2.Options),
) (*ec2.AuthorizeSecurityGroupIngressOutput, error) {
	m.RegisterCall(ctx, params)
	return m.authorizeSecurityGroupIngressOutput, m.authorizeSecurityGroupIngressError
}

// Tag methods.
func (m *ec2ServiceMock) CreateTags(
	ctx context.Context,
	params *ec2.CreateTagsInput,
	optFns ...func(*ec2.Options),
) (*ec2.CreateTagsOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createTagsOutput, m.createTagsError
}

func (m *ec2ServiceMock) DeleteTags(
	ctx context.Context,
	params *ec2.DeleteTagsInput,
	optFns ...func(*ec2.Options),
) (*ec2.DeleteTagsOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteTagsOutput, m.deleteTagsError
}
