package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cfntypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

type fakeRegistry struct {
	pages     map[cfntypes.ProvisioningType][][]string
	schemas   map[string]string
	described []string
}

func (f *fakeRegistry) ListTypes(
	_ context.Context,
	in *cloudformation.ListTypesInput,
	_ ...func(*cloudformation.Options),
) (*cloudformation.ListTypesOutput, error) {
	pages := f.pages[in.ProvisioningType]
	index := 0
	if in.NextToken != nil {
		index = int((*in.NextToken)[0] - '0')
	}
	var summaries []cfntypes.TypeSummary
	for _, name := range pages[index] {
		summaries = append(summaries, cfntypes.TypeSummary{TypeName: aws.String(name)})
	}
	var next *string
	if index+1 < len(pages) {
		next = aws.String(string(rune('0' + index + 1)))
	}
	return &cloudformation.ListTypesOutput{TypeSummaries: summaries, NextToken: next}, nil
}

func (f *fakeRegistry) DescribeType(
	_ context.Context,
	in *cloudformation.DescribeTypeInput,
	_ ...func(*cloudformation.Options),
) (*cloudformation.DescribeTypeOutput, error) {
	f.described = append(f.described, aws.ToString(in.TypeName))
	return &cloudformation.DescribeTypeOutput{Schema: aws.String(f.schemas[aws.ToString(in.TypeName)])}, nil
}

func TestDiscoverServiceTypes(t *testing.T) {
	fake := &fakeRegistry{
		pages: map[cfntypes.ProvisioningType][][]string{
			// Paginated, and AWS::SQS::Queue appears under both provisioning
			// types to exercise de-duplication.
			cfntypes.ProvisioningTypeFullyMutable: {
				{"AWS::SQS::Queue", "AWS::SQS::QueuePolicy"},
				{"AWS::SQS::QueueInlinePolicy"},
			},
			cfntypes.ProvisioningTypeImmutable: {
				{"AWS::SQS::Queue"},
			},
		},
	}

	svc := serviceEntry{Name: "SQS", Exclude: []string{"AWS::SQS::QueueInlinePolicy"}}
	got, err := discoverServiceTypes(context.Background(), fake, svc)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"AWS::SQS::Queue", "AWS::SQS::QueuePolicy"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSyncType_validatesAndWrites(t *testing.T) {
	dir := t.TempDir()
	valid := `{"typeName":"AWS::SQS::Queue","properties":{"QueueName":{"type":"string"}}}`
	fake := &fakeRegistry{schemas: map[string]string{"AWS::SQS::Queue": valid}}

	if err := syncType(context.Background(), fake, dir, "AWS::SQS::Queue"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "aws-sqs-queue.json")); err != nil {
		t.Fatalf("expected vendored schema file: %v", err)
	}

	// A malformed schema must fail rather than land in the vendored set.
	bad := &fakeRegistry{schemas: map[string]string{"AWS::Bad::Type": "{ truncated"}}
	if err := syncType(context.Background(), bad, dir, "AWS::Bad::Type"); err == nil {
		t.Error("expected error for malformed schema")
	}
	if _, err := os.Stat(filepath.Join(dir, "aws-bad-type.json")); !os.IsNotExist(err) {
		t.Error("malformed schema should not have been written")
	}
}
