package main

import (
	"fmt"
	"go/format"
	"sort"
	"strings"
)

var dataSourceSpecTypeConstants = map[string]string{
	"string":  "provider.DataSourceSpecTypeString",
	"integer": "provider.DataSourceSpecTypeInteger",
	"float":   "provider.DataSourceSpecTypeFloat",
	"boolean": "provider.DataSourceSpecTypeBoolean",
	"array":   "provider.DataSourceSpecTypeArray",
}

// Tthe operator set exposed on string filter fields, enabling
// the framework's advanced filtering (evaluated client-side by the engine).
var stringFilterOperators = []string{
	"schema.DataSourceFilterOperatorEquals",
	"schema.DataSourceFilterOperatorNotEquals",
	"schema.DataSourceFilterOperatorIn",
	"schema.DataSourceFilterOperatorNotIn",
	"schema.DataSourceFilterOperatorContains",
	"schema.DataSourceFilterOperatorStartsWith",
	"schema.DataSourceFilterOperatorEndsWith",
}

// Renders the per-type Go source for a Cloud Control–backed data
// source: a constructor (reusing the resource's schema builder and meta) plus filter
// and export field builders.
func emitDataSourceFile(ds *irDataSource, res *irResource) ([]byte, error) {
	ident := goIdentifier(res.CFNType)
	var b strings.Builder

	fmt.Fprintf(&b, "func %sDataSource(\n", ident)
	b.WriteString("\tcloudControlServiceFactory pluginutils.ServiceFactory[*aws.Config, cloudcontrolservice.Service],\n")
	b.WriteString("\tawsConfigStore pluginutils.ServiceConfigStore[*aws.Config],\n")
	b.WriteString(") provider.DataSource {\n")
	b.WriteString("\treturn cloudcontrol.CCDataSource(\n")
	b.WriteString("\t\tcloudcontrol.CCDataSourceConfig{\n")
	fmt.Fprintf(&b, "\t\t\tBlueprintType:        %q,\n", ds.BlueprintType)
	fmt.Fprintf(&b, "\t\t\tCFNType:              %q,\n", ds.CFNType)
	fmt.Fprintf(&b, "\t\t\tLabel:                %q,\n", ds.Label)
	fmt.Fprintf(&b, "\t\t\tFormattedDescription: %q,\n", ds.Description)
	fmt.Fprintf(&b, "\t\t\tSchema:               overlays.Apply(%q, %sSchema()),\n", ds.BlueprintType, ident)
	fmt.Fprintf(&b, "\t\t\tMeta: %s,\n", metaLiteral(res))
	if ds.DeriveIdentifierFromARN {
		b.WriteString("\t\t\tDeriveIdentifierFromARN: true,\n")
	}
	fmt.Fprintf(&b, "\t\t\tFilterFields:         %sDataSourceFilterFields(),\n", ident)
	fmt.Fprintf(&b, "\t\t\tExportFields:         %sDataSourceExportFields(),\n", ident)
	fmt.Fprintf(&b, "\t\t\tFormattedExamples:    readDataSourceExamples(%q),\n", exampleStem(ds.CFNType))
	b.WriteString("\t\t},\n")
	b.WriteString("\t\tcloudControlServiceFactory,\n")
	b.WriteString("\t\tawsConfigStore,\n")
	b.WriteString("\t)\n}\n\n")

	writeFilterFields(&b, ident, ds)
	writeExportFields(&b, ident, ds)

	source := fileHeader(dataSourceImportPaths()) + b.String()
	return format.Source([]byte(source))
}

func writeFilterFields(b *strings.Builder, ident string, ds *irDataSource) {
	fmt.Fprintf(b, "func %sDataSourceFilterFields() map[string]*provider.DataSourceFilterSchema {\n", ident)
	b.WriteString("\treturn map[string]*provider.DataSourceFilterSchema{\n")
	fields := append([]string(nil), ds.FilterFields...)
	sort.Strings(fields)
	for _, field := range fields {
		b.WriteString("\t\t" + fmt.Sprintf("%q", field) + ": {\n")
		b.WriteString("\t\t\tType: provider.DataSourceFilterSearchValueTypeString,\n")
		b.WriteString("\t\t\tSupportedOperators: []schema.DataSourceFilterOperator{\n")
		for _, op := range filterOperatorsFor(field) {
			fmt.Fprintf(b, "\t\t\t\t%s,\n", op)
		}
		b.WriteString("\t\t\t},\n")
		b.WriteString("\t\t},\n")
	}
	b.WriteString("\t}\n}\n\n")
}

func filterOperatorsFor(field string) []string {
	if field == "region" {
		return []string{"schema.DataSourceFilterOperatorEquals"}
	}
	return stringFilterOperators
}

func writeExportFields(b *strings.Builder, ident string, ds *irDataSource) {
	fmt.Fprintf(b, "func %sDataSourceExportFields() map[string]*provider.DataSourceSpecSchema {\n", ident)
	b.WriteString("\treturn map[string]*provider.DataSourceSpecSchema{\n")
	for _, field := range ds.ExportFields {
		fmt.Fprintf(b, "\t\t%q: {Type: %s},\n", field.Name, dataSourceSpecTypeConstants[field.Type])
	}
	b.WriteString("\t}\n}\n")
}

func dataSourceImportPaths() []string {
	return []string{
		`"github.com/aws/aws-sdk-go-v2/aws"`,
		`"github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol"`,
		`"github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/overlays"`,
		`cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"`,
		`"github.com/newstack-cloud/bluelink/libs/blueprint/provider"`,
		`"github.com/newstack-cloud/bluelink/libs/blueprint/schema"`,
		`"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"`,
	}
}

func emitDataSourceRegistryFile(dataSources []*irDataSource) ([]byte, error) {
	var b strings.Builder
	b.WriteString(fileHeader([]string{
		`"github.com/aws/aws-sdk-go-v2/aws"`,
		`cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"`,
		`"github.com/newstack-cloud/bluelink/libs/blueprint/provider"`,
		`"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"`,
	}))

	b.WriteString("// GeneratedDataSources returns the Cloud Control–backed data sources, keyed by\n")
	b.WriteString("// Bluelink data source type.\n")
	b.WriteString("func GeneratedDataSources(\n")
	b.WriteString("\tcloudControlServiceFactory pluginutils.ServiceFactory[*aws.Config, cloudcontrolservice.Service],\n")
	b.WriteString("\tawsConfigStore pluginutils.ServiceConfigStore[*aws.Config],\n")
	b.WriteString(") map[string]provider.DataSource {\n")
	b.WriteString("\treturn map[string]provider.DataSource{\n")

	sorted := append([]*irDataSource(nil), dataSources...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].BlueprintType < sorted[j].BlueprintType })
	for _, ds := range sorted {
		fmt.Fprintf(&b, "\t\t%q: %sDataSource(\n", ds.BlueprintType, goIdentifier(ds.CFNType))
		b.WriteString("\t\t\tcloudControlServiceFactory,\n")
		b.WriteString("\t\t\tawsConfigStore,\n")
		b.WriteString("\t\t),\n")
	}
	b.WriteString("\t}\n}\n")

	return format.Source([]byte(b.String()))
}
