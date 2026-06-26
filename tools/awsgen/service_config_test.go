package main

import (
	"context"
	"testing"

	cfntypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

func TestServiceEntryIncludes(t *testing.T) {
	open := serviceEntry{Name: "SQS"}
	if !open.includes("AWS::SQS::Queue") {
		t.Error("an empty Include should admit every type")
	}

	restricted := serviceEntry{Name: "EC2", Include: []string{"AWS::EC2::VPC"}}
	if !restricted.includes("AWS::EC2::VPC") {
		t.Error("listed type should be included")
	}
	if restricted.includes("AWS::EC2::Instance") {
		t.Error("unlisted type should be excluded when Include is set")
	}
}

func TestDataSourceOnlyType(t *testing.T) {
	// EC2 is configured as a data-source-only service in services.
	if !dataSourceOnlyType("AWS::EC2::VPC") {
		t.Error("EC2 should be data-source-only")
	}
	if dataSourceOnlyType("AWS::SQS::Queue") {
		t.Error("SQS resources are not data-source-only")
	}
	if dataSourceOnlyType("AWS::Unknown::Thing") {
		t.Error("unknown services are not data-source-only")
	}
}

func TestDiscoverServiceTypes_includeFilter(t *testing.T) {
	fake := &fakeRegistry{
		pages: map[cfntypes.ProvisioningType][][]string{
			cfntypes.ProvisioningTypeFullyMutable: {
				{"AWS::EC2::VPC", "AWS::EC2::Subnet", "AWS::EC2::Instance"},
				{"AWS::EC2::SecurityGroup", "AWS::EC2::FlowLog"},
			},
			cfntypes.ProvisioningTypeImmutable: {{}},
		},
	}

	svc := serviceEntry{
		Name:    "EC2",
		Include: []string{"AWS::EC2::VPC", "AWS::EC2::Subnet", "AWS::EC2::SecurityGroup"},
	}
	got, err := discoverServiceTypes(context.Background(), fake, svc)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"AWS::EC2::SecurityGroup", "AWS::EC2::Subnet", "AWS::EC2::VPC"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
