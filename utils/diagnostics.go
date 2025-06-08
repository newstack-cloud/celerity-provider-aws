package utils

import (
	"github.com/two-hundred/celerity/libs/blueprint/core"
	"github.com/two-hundred/celerity/libs/blueprint/source"
)

// GeneralDiagnosticRange returns a diagnostic range that can be used
// when the location of the diagnostic in a source config file or blueprint
// document is not known or is not applicable.
func GeneralDiagnosticRange() *core.DiagnosticRange {
	return &core.DiagnosticRange{
		Start: &source.Meta{
			Position: source.Position{
				Line:   1,
				Column: 1,
			},
		},
		End: &source.Meta{
			Position: source.Position{
				Line:   1,
				Column: 1,
			},
		},
	}
}
