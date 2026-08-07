//go:build integration

package e2e

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// A test that fails partway through leaves its whole stack standing, because the
// harness tears down only what it got as far as deploying. That is how two flex VPC
// runs left a VPC each behind, and a VPC is not a cheap thing to leak: it holds
// subnets, security groups and endpoints, and the account allows five.
//
// The sweep runs after the suite and removes what this process created and failed to
// clean up.
const (
	// Deletion is ordered so that dependencies come apart before the things holding
	// them, but AWS still needs time to catch up, most of all releasing the elastic
	// network interfaces Lambda attaches to a VPC. Those survive the function by
	// minutes, and nothing in the VPC can be deleted until they are gone.
	leakSweepENIDeadline = 12 * time.Minute
	leakSweepENIInterval = 15 * time.Second

	// AWS list APIs are eventually consistent, and SQS especially will keep returning
	// a queue for a while after it has been deleted. Reporting that as a leak would
	// fail runs that tore down perfectly well, so the sweep re-polls and only believes
	// what is still there when the account has had a chance to settle.
	leakSweepSettleTimeout  = 90 * time.Second
	leakSweepSettleInterval = 15 * time.Second

	// A hard ceiling on the whole sweep.
	//
	// The per-step deadlines above do not bound the total: waiting on network
	// interfaces is per VPC, so several leaked VPCs multiply it. More importantly,
	// nothing else bounds this. The testing package stops enforcing -timeout when
	// m.Run returns, so everything TestMain does afterwards runs unwatched, and a
	// sweep that hung would hang the run with no diagnostic at all.
	//
	// Whatever is unfinished when the budget runs out is reported and left, which is
	// the same outcome as a leak the sweep could not delete.
	leakSweepBudget = 15 * time.Minute
)

const (
	kindLambdaFunction  = "lambda function"
	kindIAMRole         = "iam role"
	kindSQSQueue        = "sqs queue"
	kindDynamoDBTable   = "dynamodb table"
	kindEventBridgeRule = "eventbridge rule"
	kindRDSCluster      = "rds db cluster"
	kindRDSSubnetGroup  = "rds db subnet group"
	kindFlexVPC         = "flex vpc"
)

// The order resources are deleted in. A kind is only reached once everything that
// could hold it has already been removed: functions before the roles they assume and
// the VPCs their network interfaces live in, RDS clusters before the subnet group
// they sit in.
var sweepDeleteOrder = []string{
	kindEventBridgeRule,
	kindLambdaFunction,
	kindRDSCluster,
	kindRDSSubnetGroup,
	kindSQSQueue,
	kindDynamoDBTable,
	kindIAMRole,
	kindFlexVPC,
}

// Each Setup registers its run-unique name prefix so the sweep can tell what this
// process created from what was already in the account.
var sweepRegistry = struct {
	mu     sync.Mutex
	scopes []string
}{}

func registerSweepScope(scope string) {
	sweepRegistry.mu.Lock()
	defer sweepRegistry.mu.Unlock()

	sweepRegistry.scopes = append(sweepRegistry.scopes, scope)
}

func registeredSweepScopes() []string {
	sweepRegistry.mu.Lock()
	defer sweepRegistry.mu.Unlock()

	return append([]string{}, sweepRegistry.scopes...)
}

// A resource left in the account after teardown should have removed it.
type leakedResource struct {
	kind string
	// What to show a human, and for most kinds what the delete call takes.
	id string
	// What the delete call needs where that differs from id, such as a VPC's own
	// identifier when the display name comes from a tag.
	handle string
}

func (l leakedResource) String() string {
	return fmt.Sprintf("%s %s", l.kind, l.id)
}

func (l leakedResource) deleteHandle() string {
	if l.handle != "" {
		return l.handle
	}

	return l.id
}

// This removes anything this process deployed and failed to tear down, and
// reports whether it found something. It is called from TestMain after the suite has
// run.
//
// Only resources whose names carry a scope registered by this process are touched. A
// stray resource that matches the naming shape but no live scope may belong to another
// developer or a concurrent run in the same account, so it is reported and left alone:
// a sweep that guesses is worse than a leak.
func runLeakSweep() bool {
	scopes := registeredSweepScopes()
	if len(scopes) == 0 {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), leakSweepBudget)
	defer cancel()

	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "leak sweep: could not load AWS config: %v\n", err)
		return false
	}

	clients := newSweepClients(cfg)
	leaks := clients.sweep(ctx, scopes)
	if len(leaks) == 0 {
		return false
	}

	fmt.Fprintf(os.Stderr, "\nleak sweep: %d resource(s) survived teardown\n", len(leaks))
	for _, leak := range leaks {
		fmt.Fprintf(os.Stderr, "  %s\n", leak)
	}

	failures := clients.deleteAll(ctx, leaks)
	for _, failure := range failures {
		fmt.Fprintf(os.Stderr, "  could not delete: %s\n", failure)
	}

	return true
}

type sweepClients struct {
	ec2          *ec2.Client
	lambda       *lambda.Client
	iam          *iam.Client
	sqs          *sqs.Client
	dynamodb     *dynamodb.Client
	eventbridge  *eventbridge.Client
	cloudcontrol *cloudcontrol.Client
}

func newSweepClients(cfg aws.Config) *sweepClients {
	return &sweepClients{
		ec2:          ec2.NewFromConfig(cfg),
		lambda:       lambda.NewFromConfig(cfg),
		iam:          iam.NewFromConfig(cfg),
		sqs:          sqs.NewFromConfig(cfg),
		dynamodb:     dynamodb.NewFromConfig(cfg),
		eventbridge:  eventbridge.NewFromConfig(cfg),
		cloudcontrol: cloudcontrol.NewFromConfig(cfg),
	}
}

// A resource the teardown deleted moments ago can still come back from a list call, so
// a single look would report leaks that are not there. Only what is still present once
// the account has settled counts.
func (c *sweepClients) sweep(ctx context.Context, scopes []string) []leakedResource {
	deadline := time.Now().Add(leakSweepSettleTimeout)
	for {
		leaks := c.sweepOnce(ctx, scopes)
		if len(leaks) == 0 {
			return nil
		}
		if time.Now().After(deadline) {
			return leaks
		}
		if err := sleepOrDone(ctx, leakSweepSettleInterval); err != nil {
			return leaks
		}
	}
}

// A sleep that gives up when the sweep's budget is spent, so no wait can outlive it.
func sleepOrDone(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (c *sweepClients) sweepOnce(ctx context.Context, scopes []string) []leakedResource {
	finders := []func(context.Context, []string) []leakedResource{
		c.leakedLambdaFunctions,
		c.leakedIAMRoles,
		c.leakedQueues,
		c.leakedTables,
		c.leakedRules,
		c.leakedRDSClusters,
		c.leakedRDSSubnetGroups,
		c.leakedFlexVPCs,
	}

	leaks := []leakedResource{}
	for _, find := range finders {
		leaks = append(leaks, find(ctx, scopes)...)
	}

	return orderForDeletion(leaks)
}

// This arranges leaks so dependencies are removed before the resources
// that hold them, and sorts within a kind so a run's output does not depend on the
// order AWS happened to list things in.
func orderForDeletion(leaks []leakedResource) []leakedResource {
	rank := map[string]int{}
	for index, kind := range sweepDeleteOrder {
		rank[kind] = index
	}

	ordered := append([]leakedResource{}, leaks...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if rank[ordered[i].kind] != rank[ordered[j].kind] {
			return rank[ordered[i].kind] < rank[ordered[j].kind]
		}
		return ordered[i].id < ordered[j].id
	})

	return ordered
}

func (c *sweepClients) deleteAll(ctx context.Context, leaks []leakedResource) []string {
	failures := []string{}
	for index, leak := range leaks {
		if ctx.Err() != nil {
			for _, remaining := range leaks[index:] {
				failures = append(
					failures,
					fmt.Sprintf("%s: sweep budget of %s spent before it was reached", remaining, leakSweepBudget),
				)
			}
			return failures
		}

		if err := c.deleteLeak(ctx, leak); err != nil && !alreadyGone(err) {
			failures = append(failures, fmt.Sprintf("%s: %v", leak, err))
		}
	}

	return failures
}

// A resource that has gone since it was listed is the outcome the sweep wanted, not a
// failure to report. Deletes race the same eventual consistency the listing does, and
// several of these APIs answer a delete of something absent with an error rather than
// treating it as a no-op.
func alreadyGone(err error) bool {
	message := strings.ToLower(err.Error())
	for _, signature := range []string{
		"nonexistent",
		"not found",
		"notfound",
		"does not exist",
		"resourcenotfoundexception",
		"nosuchentity",
		"invalidvpcid.notfound",
		"invalidgroup.notfound",
		"invalidsubnetid.notfound",
	} {
		if strings.Contains(message, signature) {
			return true
		}
	}

	return false
}

func (c *sweepClients) deleteLeak(ctx context.Context, leak leakedResource) error {
	switch leak.kind {
	case kindLambdaFunction:
		return c.deleteFunction(ctx, leak.deleteHandle())
	case kindIAMRole:
		return c.deleteRole(ctx, leak.deleteHandle())
	case kindSQSQueue:
		return c.deleteQueue(ctx, leak.deleteHandle())
	case kindDynamoDBTable:
		return c.deleteTable(ctx, leak.deleteHandle())
	case kindEventBridgeRule:
		return c.deleteRule(ctx, leak.deleteHandle())
	case kindRDSCluster:
		return c.deleteCloudControlResource(ctx, "AWS::RDS::DBCluster", leak.deleteHandle())
	case kindRDSSubnetGroup:
		return c.deleteCloudControlResource(ctx, "AWS::RDS::DBSubnetGroup", leak.deleteHandle())
	case kindFlexVPC:
		return c.deleteVPC(ctx, leak.deleteHandle())
	default:
		return fmt.Errorf("no delete implemented for %s", leak.kind)
	}
}

func nameInScopes(name string, scopes []string) bool {
	for _, scope := range scopes {
		if strings.HasPrefix(name, scope) {
			return true
		}
	}

	return false
}

func newLeak(kind string, id string) leakedResource {
	return leakedResource{kind: kind, id: id}
}

func sweepWarn(operation string, err error) {
	fmt.Fprintf(os.Stderr, "leak sweep: %s failed: %v\n", operation, err)
}
