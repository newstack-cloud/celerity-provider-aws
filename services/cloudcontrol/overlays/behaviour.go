package overlays

import (
	"context"

	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// SpecTransform mutates a spec mapping node in place. Used for the bespoke field
// handling some resources need that the generic engine cannot infer.
type SpecTransform func(spec *core.MappingNode) error

// NameGeneration auto-populates a name field with a unique provenance name when the
// user leaves it empty, reusing the shared name generators.
type NameGeneration struct {
	// Field is the camelCase spec field to populate (e.g. "queueName").
	Field string
	// Generate produces the unique name; it may read the resolved spec from the
	// input (e.g. to vary behaviour for FIFO queues).
	Generate utils.UniqueNameGenerator
}

// Behaviour carries per-type, hand-written behaviour the generic Cloud Control engine
// cannot derive from a schema: custom validation, unique-name generation, and
// transforms applied to desired/external state. It is the functionality-level equivalent to
// the schema and example overlays.
type Behaviour struct {
	CustomValidate func(
		context.Context,
		*provider.ResourceValidateInput,
	) (*provider.ResourceValidateOutput, error)
	Name                   *NameGeneration
	BeforeCreate           SpecTransform
	AfterReadExternalState SpecTransform
}

var behaviours = map[string]*Behaviour{}

// RegisterBehaviour associates a behaviour overlay with a Bluelink resource type,
// from a per-type init function. It panics on a duplicate registration.
func RegisterBehaviour(blueprintType string, behaviour *Behaviour) {
	if _, exists := behaviours[blueprintType]; exists {
		panic("duplicate behaviour overlay registered for " + blueprintType)
	}
	behaviours[blueprintType] = behaviour
}

// BehaviourFor returns the behaviour overlay for a type, or nil if none is registered.
func BehaviourFor(blueprintType string) *Behaviour {
	return behaviours[blueprintType]
}
