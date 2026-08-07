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

type Ec2ServiceMock struct {
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
	createRouteInputs            []*ec2.CreateRouteInput

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

	// Network interface-related mock fields
	describeNetworkInterfacesOutputs []*ec2.DescribeNetworkInterfacesOutput
	describeNetworkInterfacesError   error
	deleteNetworkInterfaceError      error
	deleteNetworkInterfaceInputs     []*ec2.DeleteNetworkInterfaceInput
	describeNetworkInterfacesCalls   int

	// Managed prefix list-related mock fields
	describeManagedPrefixListsOutput *ec2.DescribeManagedPrefixListsOutput
	describeManagedPrefixListsError  error

	// Egress-only internet gateway-related mock fields
	createEgressOnlyInternetGatewayOutput    *ec2.CreateEgressOnlyInternetGatewayOutput
	createEgressOnlyInternetGatewayError     error
	describeEgressOnlyInternetGatewaysOutput *ec2.DescribeEgressOnlyInternetGatewaysOutput
	describeEgressOnlyInternetGatewaysError  error
	deleteEgressOnlyInternetGatewayOutput    *ec2.DeleteEgressOnlyInternetGatewayOutput
	deleteEgressOnlyInternetGatewayError     error

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
	describeSecurityGroupsOutputs   []*ec2.DescribeSecurityGroupsOutput
	describeSecurityGroupsCalls     int
	describeSecurityGroupsError     error
	deletedSecurityGroupIDs         []string
	callOrder                       int
	lastRevokeOrder                 int
	firstDeleteSecurityGroupOrder   int
	deleteSecurityGroupOutput       *ec2.DeleteSecurityGroupOutput
	deleteSecurityGroupError        error
	deleteSecurityGroupErrorSeq     []error
	deleteSecurityGroupCallCount    int
	revokeSecurityGroupEgressOutput *ec2.RevokeSecurityGroupEgressOutput
	revokeSecurityGroupEgressError  error

	revokeSecurityGroupIngressOutput *ec2.RevokeSecurityGroupIngressOutput
	revokeSecurityGroupIngressError  error

	describeSecurityGroupRulesOutput *ec2.DescribeSecurityGroupRulesOutput
	describeSecurityGroupRulesError  error

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

	authorizeSecurityGroupEgressOutput *ec2.AuthorizeSecurityGroupEgressOutput
	authorizeSecurityGroupEgressError  error

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

type ec2ServiceMockOption func(*Ec2ServiceMock)

func CreateEc2ServiceMockFactory(
	opts ...ec2ServiceMockOption,
) func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
	return func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
		return CreateEc2ServiceMock(opts...)
	}
}

func CreateEc2ServiceMock(
	opts ...ec2ServiceMockOption,
) *Ec2ServiceMock {
	mock := &Ec2ServiceMock{}
	for _, opt := range opts {
		opt(mock)
	}
	return mock
}

func NewMockService(t any) *Ec2ServiceMock {
	return &Ec2ServiceMock{}
}

// VPC-related options.
func WithCreateVpcOutput(output *ec2.CreateVpcOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.createVpcOutput = output
	}
}

func WithCreateVpcError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.createVpcError = err
	}
}

func WithDescribeVpcsOutputs(outputs []*ec2.DescribeVpcsOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeVpcsOutputs = outputs
	}
}

func WithDescribeVpcsError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeVpcsError = err
	}
}

func WithDeleteVpcOutput(output *ec2.DeleteVpcOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.deleteVpcOutput = output
	}
}

func WithDeleteVpcError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.deleteVpcError = err
	}
}

// VPC attribute-related options.
func WithModifyVpcAttributeOutput(output *ec2.ModifyVpcAttributeOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.modifyVpcAttributeOutput = output
	}
}

func WithModifyVpcAttributeError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.modifyVpcAttributeError = err
	}
}

func WithDescribeVpcAttributeOutput(output *ec2.DescribeVpcAttributeOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeVpcAttributeOutput = output
	}
}

func WithDescribeVpcAttributeError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeVpcAttributeError = err
	}
}

// Subnet-related options.
func WithCreateSubnetOutputs(outputs []*ec2.CreateSubnetOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.createSubnetOutputs = outputs
	}
}

func WithCreateSubnetError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.createSubnetError = err
	}
}

func WithDescribeSubnetsOutput(output *ec2.DescribeSubnetsOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeSubnetsOutput = output
	}
}

func WithDescribeSubnetsError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeSubnetsError = err
	}
}

func WithDeleteSubnetOutput(output *ec2.DeleteSubnetOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.deleteSubnetOutput = output
	}
}

func WithDeleteSubnetError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.deleteSubnetError = err
	}
}

func WithModifySubnetAttributeOutput(output *ec2.ModifySubnetAttributeOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.modifySubnetAttributeOutput = output
	}
}

func WithModifySubnetAttributeError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.modifySubnetAttributeError = err
	}
}

// Availability zone-related options.
func WithDescribeAvailabilityZonesOutput(output *ec2.DescribeAvailabilityZonesOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeAvailabilityZonesOutput = output
	}
}

func WithDescribeAvailabilityZonesError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeAvailabilityZonesError = err
	}
}

// Route table-related options.
func WithCreateRouteTableOutputs(outputs []*ec2.CreateRouteTableOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.createRouteTableOutputs = outputs
	}
}

func WithCreateRouteTableError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.createRouteTableError = err
	}
}

func WithDescribeRouteTablesOutput(output *ec2.DescribeRouteTablesOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeRouteTablesOutput = output
	}
}

func WithDescribeRouteTablesError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeRouteTablesError = err
	}
}

func WithDeleteRouteTableOutput(output *ec2.DeleteRouteTableOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.deleteRouteTableOutput = output
	}
}

func WithDeleteRouteTableError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.deleteRouteTableError = err
	}
}

func WithAssociateRouteTableOutput(output *ec2.AssociateRouteTableOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.associateRouteTableOutput = output
	}
}

func WithAssociateRouteTableError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.associateRouteTableError = err
	}
}

func WithDisassociateRouteTableOutput(output *ec2.DisassociateRouteTableOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.disassociateRouteTableOutput = output
	}
}

func WithDisassociateRouteTableError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.disassociateRouteTableError = err
	}
}

func WithCreateRouteOutput(output *ec2.CreateRouteOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.createRouteOutput = output
	}
}

func WithCreateRouteError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.createRouteError = err
	}
}

// Internet gateway-related options.
func WithCreateInternetGatewayOutput(output *ec2.CreateInternetGatewayOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.createInternetGatewayOutput = output
	}
}

func WithCreateInternetGatewayError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.createInternetGatewayError = err
	}
}

// Network interface-related options.

// WithDescribeNetworkInterfacesOutputs answers successive describes in order, so a test
// can model interfaces detaching between polls.
func WithDescribeNetworkInterfacesOutputs(outputs []*ec2.DescribeNetworkInterfacesOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeNetworkInterfacesOutputs = outputs
	}
}

func WithDeleteNetworkInterfaceError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.deleteNetworkInterfaceError = err
	}
}

// Managed prefix list-related options.
func WithDescribeManagedPrefixListsOutput(output *ec2.DescribeManagedPrefixListsOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeManagedPrefixListsOutput = output
	}
}

func WithDescribeManagedPrefixListsError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeManagedPrefixListsError = err
	}
}

// Egress-only internet gateway-related options.
func WithCreateEgressOnlyInternetGatewayOutput(output *ec2.CreateEgressOnlyInternetGatewayOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.createEgressOnlyInternetGatewayOutput = output
	}
}

func WithCreateEgressOnlyInternetGatewayError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.createEgressOnlyInternetGatewayError = err
	}
}

func WithDescribeEgressOnlyInternetGatewaysOutput(output *ec2.DescribeEgressOnlyInternetGatewaysOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeEgressOnlyInternetGatewaysOutput = output
	}
}

func WithDescribeEgressOnlyInternetGatewaysError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeEgressOnlyInternetGatewaysError = err
	}
}

func WithDeleteEgressOnlyInternetGatewayOutput(output *ec2.DeleteEgressOnlyInternetGatewayOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.deleteEgressOnlyInternetGatewayOutput = output
	}
}

func WithDeleteEgressOnlyInternetGatewayError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.deleteEgressOnlyInternetGatewayError = err
	}
}

func WithDescribeInternetGatewaysOutput(output *ec2.DescribeInternetGatewaysOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeInternetGatewaysOutput = output
	}
}

func WithDescribeInternetGatewaysError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeInternetGatewaysError = err
	}
}

func WithDeleteInternetGatewayOutput(output *ec2.DeleteInternetGatewayOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.deleteInternetGatewayOutput = output
	}
}

func WithDeleteInternetGatewayError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.deleteInternetGatewayError = err
	}
}

func WithAttachInternetGatewayOutput(output *ec2.AttachInternetGatewayOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.attachInternetGatewayOutput = output
	}
}

func WithAttachInternetGatewayError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.attachInternetGatewayError = err
	}
}

func WithDetachInternetGatewayOutput(output *ec2.DetachInternetGatewayOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.detachInternetGatewayOutput = output
	}
}

func WithDetachInternetGatewayError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.detachInternetGatewayError = err
	}
}

// NAT gateway-related options.
func WithCreateNatGatewayOutputs(outputs []*ec2.CreateNatGatewayOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.createNatGatewayOutputs = outputs
	}
}

func WithCreateNatGatewayError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.createNatGatewayError = err
	}
}

func WithDescribeNatGatewaysOutput(output *ec2.DescribeNatGatewaysOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeNatGatewaysOutput = output
	}
}

func WithDescribeNatGatewaysError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeNatGatewaysError = err
	}
}

func WithDeleteNatGatewayOutput(output *ec2.DeleteNatGatewayOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.deleteNatGatewayOutput = output
	}
}

func WithDeleteNatGatewayError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.deleteNatGatewayError = err
	}
}

// Elastic IP-related options.
func WithAllocateAddressOutputs(outputs []*ec2.AllocateAddressOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.allocateAddressOutputs = outputs
	}
}

func WithAllocateAddressError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.allocateAddressError = err
	}
}

func WithReleaseAddressOutput(output *ec2.ReleaseAddressOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.releaseAddressOutput = output
	}
}

func WithReleaseAddressError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.releaseAddressError = err
	}
}

// Security group-related options.
func WithCreateSecurityGroupOutput(output *ec2.CreateSecurityGroupOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.createSecurityGroupOutput = output
	}
}

func WithCreateSecurityGroupError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.createSecurityGroupError = err
	}
}

func WithDescribeSecurityGroupsOutput(output *ec2.DescribeSecurityGroupsOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeSecurityGroupsOutput = output
	}
}

func WithDescribeSecurityGroupsError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeSecurityGroupsError = err
	}
}

// WithDeleteSecurityGroupErrorsThenSuccess fails the first N delete attempts with the
// given errors, then succeeds. Models AWS releasing a dependency after a delay.
func WithDeleteSecurityGroupErrorsThenSuccess(errs []error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.deleteSecurityGroupErrorSeq = errs
		m.deleteSecurityGroupOutput = &ec2.DeleteSecurityGroupOutput{}
	}
}

func WithDeleteSecurityGroupOutput(output *ec2.DeleteSecurityGroupOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.deleteSecurityGroupOutput = output
	}
}

func WithDeleteSecurityGroupError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.deleteSecurityGroupError = err
	}
}

func WithRevokeSecurityGroupEgressOutput(output *ec2.RevokeSecurityGroupEgressOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.revokeSecurityGroupEgressOutput = output
	}
}

func WithRevokeSecurityGroupEgressError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.revokeSecurityGroupEgressError = err
	}
}

func WithRevokeSecurityGroupIngressOutput(output *ec2.RevokeSecurityGroupIngressOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.revokeSecurityGroupIngressOutput = output
	}
}

func WithRevokeSecurityGroupIngressError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.revokeSecurityGroupIngressError = err
	}
}

func WithDescribeSecurityGroupRulesOutput(
	output *ec2.DescribeSecurityGroupRulesOutput,
) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeSecurityGroupRulesOutput = output
	}
}

func WithDescribeSecurityGroupRulesError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeSecurityGroupRulesError = err
	}
}

// Network ACL-related options.
func WithCreateNetworkAclOutput(output *ec2.CreateNetworkAclOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.createNetworkAclOutput = output
	}
}

func WithCreateNetworkAclError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.createNetworkAclError = err
	}
}

func WithDescribeNetworkAclsOutput(output *ec2.DescribeNetworkAclsOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeNetworkAclsOutput = output
	}
}

func WithDescribeNetworkAclsError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeNetworkAclsError = err
	}
}

func WithDeleteNetworkAclOutput(output *ec2.DeleteNetworkAclOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.deleteNetworkAclOutput = output
	}
}

func WithDeleteNetworkAclError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.deleteNetworkAclError = err
	}
}

func WithCreateNetworkAclEntryOutput(output *ec2.CreateNetworkAclEntryOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.createNetworkAclEntryOutput = output
	}
}

func WithCreateNetworkAclEntryError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.createNetworkAclEntryError = err
	}
}

func WithReplaceNetworkAclAssociationOutput(output *ec2.ReplaceNetworkAclAssociationOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.replaceNetworkAclAssociationOutput = output
	}
}

func WithReplaceNetworkAclAssociationError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.replaceNetworkAclAssociationError = err
	}
}

// VPC endpoint-related options.
func WithDescribeVpcEndpointsOutput(output *ec2.DescribeVpcEndpointsOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeVpcEndpointsOutput = output
	}
}

func WithDescribeVpcEndpointsError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeVpcEndpointsError = err
	}
}

func WithCreateVpcEndpointOutput(output *ec2.CreateVpcEndpointOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.createVpcEndpointOutput = output
	}
}

func WithCreateVpcEndpointError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.createVpcEndpointError = err
	}
}

func WithModifyVpcEndpointOutput(output *ec2.ModifyVpcEndpointOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.modifyVpcEndpointOutput = output
	}
}

func WithModifyVpcEndpointError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.modifyVpcEndpointError = err
	}
}

func WithDeleteVpcEndpointsOutput(output *ec2.DeleteVpcEndpointsOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.deleteVpcEndpointsOutput = output
	}
}

func WithDeleteVpcEndpointsError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.deleteVpcEndpointsError = err
	}
}

// Security group ingress-related options.
func WithAuthorizeSecurityGroupIngressOutput(output *ec2.AuthorizeSecurityGroupIngressOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.authorizeSecurityGroupIngressOutput = output
	}
}

func WithAuthorizeSecurityGroupEgressOutput(output *ec2.AuthorizeSecurityGroupEgressOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.authorizeSecurityGroupEgressOutput = output
	}
}

func WithAuthorizeSecurityGroupEgressError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.authorizeSecurityGroupEgressError = err
	}
}

func WithAuthorizeSecurityGroupIngressError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.authorizeSecurityGroupIngressError = err
	}
}

// Tag-related options.
func WithCreateTagsOutput(output *ec2.CreateTagsOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.createTagsOutput = output
	}
}

func WithCreateTagsError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.createTagsError = err
	}
}

func WithDeleteTagsOutput(output *ec2.DeleteTagsOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.deleteTagsOutput = output
	}
}

func WithDeleteTagsError(err error) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.deleteTagsError = err
	}
}

// VPC methods.
func (m *Ec2ServiceMock) CreateVpc(
	ctx context.Context,
	params *ec2.CreateVpcInput,
	optFns ...func(*ec2.Options),
) (*ec2.CreateVpcOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createVpcOutput, m.createVpcError
}

func (m *Ec2ServiceMock) DescribeVpcs(
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

func (m *Ec2ServiceMock) DeleteVpc(
	ctx context.Context,
	params *ec2.DeleteVpcInput,
	optFns ...func(*ec2.Options),
) (*ec2.DeleteVpcOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteVpcOutput, m.deleteVpcError
}

func (m *Ec2ServiceMock) ModifyVpcAttribute(
	ctx context.Context,
	params *ec2.ModifyVpcAttributeInput,
	optFns ...func(*ec2.Options),
) (*ec2.ModifyVpcAttributeOutput, error) {
	m.RegisterCall(ctx, params)
	return m.modifyVpcAttributeOutput, m.modifyVpcAttributeError
}

func (m *Ec2ServiceMock) DescribeVpcAttribute(
	ctx context.Context,
	params *ec2.DescribeVpcAttributeInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeVpcAttributeOutput, error) {
	m.RegisterCall(ctx, params)
	return m.describeVpcAttributeOutput, m.describeVpcAttributeError
}

// Subnet methods.
func (m *Ec2ServiceMock) CreateSubnet(
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

func (m *Ec2ServiceMock) DescribeSubnets(
	ctx context.Context,
	params *ec2.DescribeSubnetsInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeSubnetsOutput, error) {
	m.RegisterCall(ctx, params)
	return m.describeSubnetsOutput, m.describeSubnetsError
}

func (m *Ec2ServiceMock) DeleteSubnet(
	ctx context.Context,
	params *ec2.DeleteSubnetInput,
	optFns ...func(*ec2.Options),
) (*ec2.DeleteSubnetOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteSubnetOutput, m.deleteSubnetError
}

func (m *Ec2ServiceMock) ModifySubnetAttribute(
	ctx context.Context,
	params *ec2.ModifySubnetAttributeInput,
	optFns ...func(*ec2.Options),
) (*ec2.ModifySubnetAttributeOutput, error) {
	m.RegisterCall(ctx, params)
	return m.modifySubnetAttributeOutput, m.modifySubnetAttributeError
}

// Availability zone methods.
func (m *Ec2ServiceMock) DescribeAvailabilityZones(
	ctx context.Context,
	params *ec2.DescribeAvailabilityZonesInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeAvailabilityZonesOutput, error) {
	m.RegisterCall(ctx, params)
	return m.describeAvailabilityZonesOutput, m.describeAvailabilityZonesError
}

// Route table methods.
func (m *Ec2ServiceMock) CreateRouteTable(
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

func (m *Ec2ServiceMock) DescribeRouteTables(
	ctx context.Context,
	params *ec2.DescribeRouteTablesInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeRouteTablesOutput, error) {
	m.RegisterCall(ctx, params)
	return m.describeRouteTablesOutput, m.describeRouteTablesError
}

func (m *Ec2ServiceMock) DeleteRouteTable(
	ctx context.Context,
	params *ec2.DeleteRouteTableInput,
	optFns ...func(*ec2.Options),
) (*ec2.DeleteRouteTableOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteRouteTableOutput, m.deleteRouteTableError
}

func (m *Ec2ServiceMock) AssociateRouteTable(
	ctx context.Context,
	params *ec2.AssociateRouteTableInput,
	optFns ...func(*ec2.Options),
) (*ec2.AssociateRouteTableOutput, error) {
	m.RegisterCall(ctx, params)
	return m.associateRouteTableOutput, m.associateRouteTableError
}

func (m *Ec2ServiceMock) DisassociateRouteTable(
	ctx context.Context,
	params *ec2.DisassociateRouteTableInput,
	optFns ...func(*ec2.Options),
) (*ec2.DisassociateRouteTableOutput, error) {
	m.RegisterCall(ctx, params)
	return m.disassociateRouteTableOutput, m.disassociateRouteTableError
}

func (m *Ec2ServiceMock) CreateRoute(
	ctx context.Context,
	params *ec2.CreateRouteInput,
	optFns ...func(*ec2.Options),
) (*ec2.CreateRouteOutput, error) {
	m.mu.Lock()
	m.createRouteInputs = append(m.createRouteInputs, params)
	m.mu.Unlock()

	m.RegisterCall(ctx, params)
	return m.createRouteOutput, m.createRouteError
}

// CreateRouteInputs returns every route the caller asked for, in order.
//
// MockCalls can only assert on a call at a known index, and a VPC creates one route
// per subnet per address family, so pinning an index would couple the assertion to
// the order subnets happen to be iterated in. Tests that care which gateway a route
// points at read the inputs instead.
func (m *Ec2ServiceMock) CreateRouteInputs() []*ec2.CreateRouteInput {
	m.mu.Lock()
	defer m.mu.Unlock()

	return append([]*ec2.CreateRouteInput{}, m.createRouteInputs...)
}

// Internet gateway methods.
func (m *Ec2ServiceMock) CreateInternetGateway(
	ctx context.Context,
	params *ec2.CreateInternetGatewayInput,
	optFns ...func(*ec2.Options),
) (*ec2.CreateInternetGatewayOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createInternetGatewayOutput, m.createInternetGatewayError
}

func (m *Ec2ServiceMock) DescribeNetworkInterfaces(
	ctx context.Context,
	params *ec2.DescribeNetworkInterfacesInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeNetworkInterfacesOutput, error) {
	m.RegisterCall(ctx, params)

	m.mu.Lock()
	call := m.describeNetworkInterfacesCalls
	m.describeNetworkInterfacesCalls++
	outputs := m.describeNetworkInterfacesOutputs
	m.mu.Unlock()

	if len(outputs) == 0 {
		return &ec2.DescribeNetworkInterfacesOutput{}, m.describeNetworkInterfacesError
	}
	if call < len(outputs) {
		return outputs[call], m.describeNetworkInterfacesError
	}

	return outputs[len(outputs)-1], m.describeNetworkInterfacesError
}

func (m *Ec2ServiceMock) DeleteNetworkInterface(
	ctx context.Context,
	params *ec2.DeleteNetworkInterfaceInput,
	optFns ...func(*ec2.Options),
) (*ec2.DeleteNetworkInterfaceOutput, error) {
	m.RegisterCall(ctx, params)

	m.mu.Lock()
	m.deleteNetworkInterfaceInputs = append(m.deleteNetworkInterfaceInputs, params)
	m.mu.Unlock()

	return &ec2.DeleteNetworkInterfaceOutput{}, m.deleteNetworkInterfaceError
}

// DeletedNetworkInterfaceIDs reports the interfaces the caller reaped, in order.
func (m *Ec2ServiceMock) DeletedNetworkInterfaceIDs() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	ids := make([]string, 0, len(m.deleteNetworkInterfaceInputs))
	for _, input := range m.deleteNetworkInterfaceInputs {
		ids = append(ids, aws.ToString(input.NetworkInterfaceId))
	}

	return ids
}

func (m *Ec2ServiceMock) DescribeManagedPrefixLists(
	ctx context.Context,
	params *ec2.DescribeManagedPrefixListsInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeManagedPrefixListsOutput, error) {
	m.RegisterCall(ctx, params)
	return m.describeManagedPrefixListsOutput, m.describeManagedPrefixListsError
}

func (m *Ec2ServiceMock) CreateEgressOnlyInternetGateway(
	ctx context.Context,
	params *ec2.CreateEgressOnlyInternetGatewayInput,
	optFns ...func(*ec2.Options),
) (*ec2.CreateEgressOnlyInternetGatewayOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createEgressOnlyInternetGatewayOutput, m.createEgressOnlyInternetGatewayError
}

func (m *Ec2ServiceMock) DescribeEgressOnlyInternetGateways(
	ctx context.Context,
	params *ec2.DescribeEgressOnlyInternetGatewaysInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeEgressOnlyInternetGatewaysOutput, error) {
	m.RegisterCall(ctx, params)
	return m.describeEgressOnlyInternetGatewaysOutput, m.describeEgressOnlyInternetGatewaysError
}

func (m *Ec2ServiceMock) DeleteEgressOnlyInternetGateway(
	ctx context.Context,
	params *ec2.DeleteEgressOnlyInternetGatewayInput,
	optFns ...func(*ec2.Options),
) (*ec2.DeleteEgressOnlyInternetGatewayOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteEgressOnlyInternetGatewayOutput, m.deleteEgressOnlyInternetGatewayError
}

func (m *Ec2ServiceMock) DescribeInternetGateways(
	ctx context.Context,
	params *ec2.DescribeInternetGatewaysInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeInternetGatewaysOutput, error) {
	m.RegisterCall(ctx, params)
	return m.describeInternetGatewaysOutput, m.describeInternetGatewaysError
}

func (m *Ec2ServiceMock) DeleteInternetGateway(
	ctx context.Context,
	params *ec2.DeleteInternetGatewayInput,
	optFns ...func(*ec2.Options),
) (*ec2.DeleteInternetGatewayOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteInternetGatewayOutput, m.deleteInternetGatewayError
}

func (m *Ec2ServiceMock) AttachInternetGateway(
	ctx context.Context,
	params *ec2.AttachInternetGatewayInput,
	optFns ...func(*ec2.Options),
) (*ec2.AttachInternetGatewayOutput, error) {
	m.RegisterCall(ctx, params)
	return m.attachInternetGatewayOutput, m.attachInternetGatewayError
}

func (m *Ec2ServiceMock) DetachInternetGateway(
	ctx context.Context,
	params *ec2.DetachInternetGatewayInput,
	optFns ...func(*ec2.Options),
) (*ec2.DetachInternetGatewayOutput, error) {
	m.RegisterCall(ctx, params)
	return m.detachInternetGatewayOutput, m.detachInternetGatewayError
}

// NAT gateway methods.
func (m *Ec2ServiceMock) CreateNatGateway(
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

func (m *Ec2ServiceMock) DescribeNatGateways(
	ctx context.Context,
	params *ec2.DescribeNatGatewaysInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeNatGatewaysOutput, error) {
	m.RegisterCall(ctx, params)
	return m.describeNatGatewaysOutput, m.describeNatGatewaysError
}

func (m *Ec2ServiceMock) DeleteNatGateway(
	ctx context.Context,
	params *ec2.DeleteNatGatewayInput,
	optFns ...func(*ec2.Options),
) (*ec2.DeleteNatGatewayOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteNatGatewayOutput, m.deleteNatGatewayError
}

// Elastic IP methods.
func (m *Ec2ServiceMock) AllocateAddress(
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

func (m *Ec2ServiceMock) ReleaseAddress(
	ctx context.Context,
	params *ec2.ReleaseAddressInput,
	optFns ...func(*ec2.Options),
) (*ec2.ReleaseAddressOutput, error) {
	m.RegisterCall(ctx, params)
	return m.releaseAddressOutput, m.releaseAddressError
}

// Security group methods.
func (m *Ec2ServiceMock) CreateSecurityGroup(
	ctx context.Context,
	params *ec2.CreateSecurityGroupInput,
	optFns ...func(*ec2.Options),
) (*ec2.CreateSecurityGroupOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createSecurityGroupOutput, m.createSecurityGroupError
}

func (m *Ec2ServiceMock) DescribeSecurityGroups(
	ctx context.Context,
	params *ec2.DescribeSecurityGroupsInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeSecurityGroupsOutput, error) {
	m.RegisterCall(ctx, params)

	m.mu.Lock()
	call := m.describeSecurityGroupsCalls
	m.describeSecurityGroupsCalls++
	sequenced := m.describeSecurityGroupsOutputs
	m.mu.Unlock()

	if len(sequenced) > 0 {
		if call < len(sequenced) {
			return sequenced[call], m.describeSecurityGroupsError
		}
		return sequenced[len(sequenced)-1], m.describeSecurityGroupsError
	}

	return m.describeSecurityGroupsOutput, m.describeSecurityGroupsError
}

// WithDescribeSecurityGroupsOutputs answers successive describes in order, for callers
// that look groups up more than once in a single operation.
func WithDescribeSecurityGroupsOutputs(outputs []*ec2.DescribeSecurityGroupsOutput) ec2ServiceMockOption {
	return func(m *Ec2ServiceMock) {
		m.describeSecurityGroupsOutputs = outputs
	}
}

// DeletedSecurityGroupIDs reports the groups the caller deleted, in order.
func (m *Ec2ServiceMock) DeletedSecurityGroupIDs() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	return append([]string{}, m.deletedSecurityGroupIDs...)
}

// LastRevokeCallOrder and FirstDeleteSecurityGroupCallOrder let a test assert that all
// rules were revoked before any group was deleted, which is what stops groups that
// reference each other blocking one another.
func (m *Ec2ServiceMock) LastRevokeCallOrder() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.lastRevokeOrder
}

func (m *Ec2ServiceMock) FirstDeleteSecurityGroupCallOrder() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.firstDeleteSecurityGroupOrder == 0 {
		return 1 << 30
	}

	return m.firstDeleteSecurityGroupOrder
}

func (m *Ec2ServiceMock) DeleteSecurityGroup(
	ctx context.Context,
	params *ec2.DeleteSecurityGroupInput,
	optFns ...func(*ec2.Options),
) (*ec2.DeleteSecurityGroupOutput, error) {
	m.RegisterCall(ctx, params)

	m.mu.Lock()
	call := m.deleteSecurityGroupCallCount
	m.deleteSecurityGroupCallCount++
	seq := m.deleteSecurityGroupErrorSeq
	m.callOrder++
	if m.firstDeleteSecurityGroupOrder == 0 {
		m.firstDeleteSecurityGroupOrder = m.callOrder
	}
	m.deletedSecurityGroupIDs = append(m.deletedSecurityGroupIDs, aws.ToString(params.GroupId))
	m.mu.Unlock()

	if call < len(seq) {
		return m.deleteSecurityGroupOutput, seq[call]
	}

	return m.deleteSecurityGroupOutput, m.deleteSecurityGroupError
}

// DeleteSecurityGroupCallCount reports how many delete attempts were made, which is
// what distinguishes a caller that waits for a group to be released from one that
// gives up after the first failure.
func (m *Ec2ServiceMock) DeleteSecurityGroupCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.deleteSecurityGroupCallCount
}

func (m *Ec2ServiceMock) RevokeSecurityGroupEgress(
	ctx context.Context,
	params *ec2.RevokeSecurityGroupEgressInput,
	optFns ...func(*ec2.Options),
) (*ec2.RevokeSecurityGroupEgressOutput, error) {
	m.RegisterCall(ctx, params)

	m.mu.Lock()
	m.callOrder++
	m.lastRevokeOrder = m.callOrder
	m.mu.Unlock()
	return m.revokeSecurityGroupEgressOutput, m.revokeSecurityGroupEgressError
}

func (m *Ec2ServiceMock) RevokeSecurityGroupIngress(
	ctx context.Context,
	params *ec2.RevokeSecurityGroupIngressInput,
	optFns ...func(*ec2.Options),
) (*ec2.RevokeSecurityGroupIngressOutput, error) {
	m.RegisterCall(ctx, params)

	m.mu.Lock()
	m.callOrder++
	m.lastRevokeOrder = m.callOrder
	m.mu.Unlock()
	return m.revokeSecurityGroupIngressOutput, m.revokeSecurityGroupIngressError
}

func (m *Ec2ServiceMock) DescribeSecurityGroupRules(
	ctx context.Context,
	params *ec2.DescribeSecurityGroupRulesInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeSecurityGroupRulesOutput, error) {
	m.RegisterCall(ctx, params)
	return m.describeSecurityGroupRulesOutput, m.describeSecurityGroupRulesError
}

// Network ACL methods.
func (m *Ec2ServiceMock) CreateNetworkAcl(
	ctx context.Context,
	params *ec2.CreateNetworkAclInput,
	optFns ...func(*ec2.Options),
) (*ec2.CreateNetworkAclOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createNetworkAclOutput, m.createNetworkAclError
}

func (m *Ec2ServiceMock) DescribeNetworkAcls(
	ctx context.Context,
	params *ec2.DescribeNetworkAclsInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeNetworkAclsOutput, error) {
	m.RegisterCall(ctx, params)
	return m.describeNetworkAclsOutput, m.describeNetworkAclsError
}

func (m *Ec2ServiceMock) DeleteNetworkAcl(
	ctx context.Context,
	params *ec2.DeleteNetworkAclInput,
	optFns ...func(*ec2.Options),
) (*ec2.DeleteNetworkAclOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteNetworkAclOutput, m.deleteNetworkAclError
}

func (m *Ec2ServiceMock) CreateNetworkAclEntry(
	ctx context.Context,
	params *ec2.CreateNetworkAclEntryInput,
	optFns ...func(*ec2.Options),
) (*ec2.CreateNetworkAclEntryOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createNetworkAclEntryOutput, m.createNetworkAclEntryError
}

func (m *Ec2ServiceMock) ReplaceNetworkAclAssociation(
	ctx context.Context,
	params *ec2.ReplaceNetworkAclAssociationInput,
	optFns ...func(*ec2.Options),
) (*ec2.ReplaceNetworkAclAssociationOutput, error) {
	m.RegisterCall(ctx, params)
	return m.replaceNetworkAclAssociationOutput, m.replaceNetworkAclAssociationError
}

// VPC endpoint methods.
func (m *Ec2ServiceMock) DescribeVpcEndpoints(
	ctx context.Context,
	params *ec2.DescribeVpcEndpointsInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeVpcEndpointsOutput, error) {
	m.RegisterCall(ctx, params)
	return m.describeVpcEndpointsOutput, m.describeVpcEndpointsError
}

func (m *Ec2ServiceMock) CreateVpcEndpoint(
	ctx context.Context,
	params *ec2.CreateVpcEndpointInput,
	optFns ...func(*ec2.Options),
) (*ec2.CreateVpcEndpointOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createVpcEndpointOutput, m.createVpcEndpointError
}

func (m *Ec2ServiceMock) ModifyVpcEndpoint(
	ctx context.Context,
	params *ec2.ModifyVpcEndpointInput,
	optFns ...func(*ec2.Options),
) (*ec2.ModifyVpcEndpointOutput, error) {
	m.RegisterCall(ctx, params)
	return m.modifyVpcEndpointOutput, m.modifyVpcEndpointError
}

func (m *Ec2ServiceMock) DeleteVpcEndpoints(
	ctx context.Context,
	params *ec2.DeleteVpcEndpointsInput,
	optFns ...func(*ec2.Options),
) (*ec2.DeleteVpcEndpointsOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteVpcEndpointsOutput, m.deleteVpcEndpointsError
}

// Security group ingress methods.
func (m *Ec2ServiceMock) AuthorizeSecurityGroupIngress(
	ctx context.Context,
	params *ec2.AuthorizeSecurityGroupIngressInput,
	optFns ...func(*ec2.Options),
) (*ec2.AuthorizeSecurityGroupIngressOutput, error) {
	m.RegisterCall(ctx, params)
	return m.authorizeSecurityGroupIngressOutput, m.authorizeSecurityGroupIngressError
}

func (m *Ec2ServiceMock) AuthorizeSecurityGroupEgress(
	ctx context.Context,
	params *ec2.AuthorizeSecurityGroupEgressInput,
	optFns ...func(*ec2.Options),
) (*ec2.AuthorizeSecurityGroupEgressOutput, error) {
	m.RegisterCall(ctx, params)
	return m.authorizeSecurityGroupEgressOutput, m.authorizeSecurityGroupEgressError
}

// Tag methods.
func (m *Ec2ServiceMock) CreateTags(
	ctx context.Context,
	params *ec2.CreateTagsInput,
	optFns ...func(*ec2.Options),
) (*ec2.CreateTagsOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createTagsOutput, m.createTagsError
}

func (m *Ec2ServiceMock) DeleteTags(
	ctx context.Context,
	params *ec2.DeleteTagsInput,
	optFns ...func(*ec2.Options),
) (*ec2.DeleteTagsOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteTagsOutput, m.deleteTagsError
}
