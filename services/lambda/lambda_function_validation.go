package lambda

import (
	"fmt"
	"strings"

	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/schema"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
)

func validateZipFileRuntime(
	path string,
	value *core.MappingNode,
	resource *schema.Resource,
) []*core.Diagnostic {
	runtime, hasRuntime := pluginutils.GetValueByPath("$.runtime", resource.Spec)
	if !hasRuntime || runtime.StringWithSubstitutions != nil {
		// When runtime is not defined, or is not yet resolved,
		// we can't validate the zip file runtime at the resource validation
		// stage. This will be re-validated before preparing the zip file
		// in the update/create stage.
		return []*core.Diagnostic{}
	}

	runtimeValue := core.StringValue(runtime)
	if !strings.HasPrefix(runtimeValue, "nodejs") &&
		!strings.HasPrefix(runtimeValue, "python") {
		return []*core.Diagnostic{
			{
				Level: core.DiagnosticLevelError,
				Message: fmt.Sprintf(
					"The %s field with inline code is only "+
						"supported for Node.js and Python runtimes.",
					path,
				),
				Range: core.DiagnosticRangeFromSourceMeta(value.SourceMeta, nil),
			},
		}
	}

	return []*core.Diagnostic{}
}
