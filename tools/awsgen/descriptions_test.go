package main

import (
	"strings"
	"testing"
)

func TestSanitiseDescriptionDropsBoilerplate(t *testing.T) {
	cases := []string{
		"Resource Type definition for AWS::Events::Archive",
		"Resource type definition for AWS::Events::EventBus",
		"Resource Type definition for AWS::Events::ApiDestination.",
		"Version: None. Resource Type definition for AWS::DynamoDB::GlobalTable",
	}
	for _, in := range cases {
		if got := sanitiseDescription(in); got != "" {
			t.Errorf("sanitiseDescription(%q) = %q, want empty", in, got)
		}
	}
}

func TestSanitiseDescriptionReplacesInlineTypeIdentifiers(t *testing.T) {
	in := "The ``AWS::DynamoDB::Table`` resource creates a table."
	want := "The DynamoDB table resource creates a table."
	if got := sanitiseDescription(in); got != want {
		t.Errorf("sanitiseDescription(%q) = %q, want %q", in, got, want)
	}
}

func TestSanitiseDescriptionStripsCloudFormationLinks(t *testing.T) {
	in := "To attach a managed policy to a group, use " +
		"[AWS::IAM::ManagedPolicy](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-iam-managedpolicy.html)."
	got := sanitiseDescription(in)
	want := "To attach a managed policy to a group, use IAM managed policy."
	if got != want {
		t.Errorf("sanitiseDescription(%q) = %q, want %q", in, got, want)
	}
}

func TestSanitiseDescriptionKeepsNonCloudFormationLinks(t *testing.T) {
	in := "For more information, see " +
		"[IAM roles](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles.html) in the *IAM User Guide*."
	got := sanitiseDescription(in)
	if !strings.Contains(got, "[IAM roles](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles.html)") {
		t.Errorf("non-CloudFormation link should be kept, got %q", got)
	}
}

func TestSanitiseDescriptionExpandsMacrosAndFraming(t *testing.T) {
	in := "CFNlong typically creates DDB tables in parallel when your CFNlong template runs. " +
		"If LAM fails, the stack operation fails."
	got := sanitiseDescription(in)
	want := "Bluelink typically creates DynamoDB tables in parallel when your Bluelink blueprint runs. " +
		"If Lambda fails, the deployment operation fails."
	if got != want {
		t.Errorf("sanitiseDescription(%q) = %q, want %q", in, got, want)
	}
}

func TestSanitiseDescriptionHandlesBareCFN(t *testing.T) {
	in := "If you don't specify a name, CFN generates a unique physical ID."
	want := "If you don't specify a name, Bluelink generates a unique physical ID."
	if got := sanitiseDescription(in); got != want {
		t.Errorf("sanitiseDescription(%q) = %q, want %q", in, got, want)
	}
}

func TestSanitiseDescriptionReplacesPseudoParameters(t *testing.T) {
	in := "The default value resolves to AWS::Region at deployment time."
	want := "The default value resolves to the region at deployment time."
	if got := sanitiseDescription(in); got != want {
		t.Errorf("sanitiseDescription(%q) = %q, want %q", in, got, want)
	}
}

// Intrinsic-function guidance has no Bluelink analogue, so the whole sentence is dropped
// rather than rewritten, while the surrounding prose is preserved verbatim.
func TestSanitiseDescriptionDropsIntrinsicFunctionGuidance(t *testing.T) {
	cases := map[string]string{
		// Fn::Join naming guidance trails a useful sentence, which survives.
		"Naming an IAM resource can cause an unrecoverable error if you reuse the same " +
			"blueprint in multiple Regions. To prevent this, we recommend using `Fn::Join` " +
			"and `AWS::Region` to create a Region-specific name, as in the following example: " +
			"`{\"Fn::Join\": [\"\", [{\"Ref\": \"AWS::Region\"}, {\"Ref\": \"MyResourceName\"}]]}`.": "Naming an IAM resource can cause an unrecoverable error if you reuse the same " +
			"blueprint in multiple Regions.",
		// Fn::GetAtt return-value note sits between the core description and a dangling
		// "For more information" sentence; both go, the core stays.
		"The Amazon Resource Name (ARN) of the secret. This parameter is a return value " +
			"that you can retrieve using the `Fn::GetAtt` intrinsic function. For more " +
			"information, see Return values.": "The Amazon Resource Name (ARN) of the secret.",
		// The Ref/DependsOn block is wholly CloudFormation authoring guidance; the orphaned
		// "This dependency..." continuation is removed with it.
		"If an external policy has a `Ref` to a role and if a resource also has a `Ref` to " +
			"the same role, add a `DependsOn` attribute to the resource. This dependency " +
			"ensures that the role's policy is available throughout the resource's lifecycle.": "",
	}
	for in, want := range cases {
		if got := sanitiseDescription(in); got != want {
			t.Errorf("sanitiseDescription(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSanitiseDescriptionKeepsAbbreviationSentences(t *testing.T) {
	in := "Identifies an alias name (e.g. `alias/aws/sqs`), key ARN, or key ID. " +
		"The value is resolved at deployment time."
	if got := sanitiseDescription(in); got != in {
		t.Errorf("sanitiseDescription(%q) = %q, want unchanged", in, got)
	}
}

func TestSanitiseDescriptionPunctuationCleanup(t *testing.T) {
	// A space left dangling before sentence punctuation by a removal is collapsed,
	// but a legitimate leading-dot token keeps its preceding space.
	cases := map[string]string{
		"The deployment package is a .zip file archive .": "The deployment package is a .zip file archive.",
		"To attach a managed policy , use the console":    "To attach a managed policy, use the console",
	}
	for in, want := range cases {
		if got := sanitiseDescription(in); got != want {
			t.Errorf("sanitiseDescription(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSanitiseDescriptionLeavesNoCloudFormationLeakage(t *testing.T) {
	inputs := []string{
		"The ``AWS::Lambda::Function`` resource creates a Lambda function. When you update a " +
			"``AWS::Lambda::Function`` resource, CFNshort calls the UpdateFunctionConfiguration API.",
		"Creates a new managed policy for your AWS-account. This operation creates a policy version " +
			"with a version identifier of ``v1``.",
		"Version: None. Resource Type definition for AWS::DynamoDB::GlobalTable",
		"The name of the bucket. We recommend using `Fn::Sub` to build a unique name. " +
			"The bucket is created in the current region.",
	}
	for _, in := range inputs {
		got := sanitiseDescription(in)
		for _, banned := range []string{"AWS::", "CloudFormation", "CFNlong", "CFNshort", "``", "Fn::", "`Ref`"} {
			if strings.Contains(got, banned) {
				t.Errorf("sanitised description %q still contains %q", got, banned)
			}
		}
	}
}

func TestFriendlyTypePhrase(t *testing.T) {
	cases := map[string]string{
		"AWS::DynamoDB::Table":        "DynamoDB table",
		"AWS::Events::EventBusPolicy": "Events event bus policy",
		"AWS::Lambda::Function":       "Lambda function",
		"`AWS::Lambda::Url`":          "Lambda url",
		"AWS::IAM::ManagedPolicy":     "IAM managed policy",
	}
	for in, want := range cases {
		if got := friendlyTypePhrase(in); got != want {
			t.Errorf("friendlyTypePhrase(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFallbackDescription(t *testing.T) {
	cases := map[string]string{
		"AWS::Events::Archive":       "Manages an Events archive.",
		"AWS::DynamoDB::GlobalTable": "Manages a DynamoDB global table.",
		"AWS::IAM::AccessKey":        "Manages an IAM access key.",
	}
	for in, want := range cases {
		if got := fallbackDescription(in); got != want {
			t.Errorf("fallbackDescription(%q) = %q, want %q", in, got, want)
		}
	}
}
