package server

import (
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"

	databasedto "github.com/cgouz/pgo-mcp/src/modules/database/dto"
	diagnosticsdto "github.com/cgouz/pgo-mcp/src/modules/diagnostics/dto"
	indexdto "github.com/cgouz/pgo-mcp/src/modules/index/dto"
	objectdto "github.com/cgouz/pgo-mcp/src/modules/object/dto"
	querydto "github.com/cgouz/pgo-mcp/src/modules/query/dto"
	schemadto "github.com/cgouz/pgo-mcp/src/modules/schema/dto"
)

func toolOutputSchemas() map[string]*jsonschema.Schema {
	return map[string]*jsonschema.Schema{
		"list_databases":           outputSchema[databasedto.ListDatabasesOutput](),
		"list_schemas":             outputSchema[schemadto.ListSchemasOutput](),
		"list_objects":             outputSchema[objectdto.ListObjectsOutput](),
		"get_object_details":       outputSchema[objectdto.GetObjectDetailsOutput](),
		"execute_sql":              outputSchema[querydto.ExecuteSQLOutput](),
		"explain_query":            outputSchema[querydto.ExplainQueryOutput](),
		"analyze_query_indexes":    outputSchema[querydto.AnalyzeQueryIndexesOutput](),
		"get_top_queries":          outputSchema[querydto.GetTopQueriesOutput](),
		"analyze_db_health":        outputSchema[diagnosticsdto.AnalyzeDBHealthOutput](),
		"analyze_workload_indexes": outputSchema[indexdto.AnalyzeWorkloadOutput](),
	}
}

func outputSchema[T any]() *jsonschema.Schema {
	schema, err := jsonschema.For[T](nil)
	if err != nil {
		panic(fmt.Sprintf("generate output schema: %v", err))
	}

	makePortable(schema)

	return schema
}

func makePortable(schema *jsonschema.Schema) {
	if schema == nil {
		return
	}

	if len(schema.Types) > 0 {
		for _, schemaType := range schema.Types {
			schema.AnyOf = append(schema.AnyOf, &jsonschema.Schema{Type: schemaType})
		}
		schema.Types = nil
	}

	for _, definition := range schema.Defs {
		makePortable(definition)
	}
	for _, definition := range schema.Definitions {
		makePortable(definition)
	}
	for _, property := range schema.Properties {
		makePortable(property)
	}
	for _, property := range schema.PatternProperties {
		makePortable(property)
	}

	makePortable(schema.Items)
	for _, item := range schema.ItemsArray {
		makePortable(item)
	}
	for _, item := range schema.PrefixItems {
		makePortable(item)
	}
	makePortable(schema.AdditionalItems)
	makePortable(schema.AdditionalProperties)
	makePortable(schema.Contains)
	makePortable(schema.PropertyNames)
	makePortable(schema.UnevaluatedItems)
	makePortable(schema.UnevaluatedProperties)

	for _, child := range schema.AllOf {
		makePortable(child)
	}
	for _, child := range schema.AnyOf {
		makePortable(child)
	}
	for _, child := range schema.OneOf {
		makePortable(child)
	}
	makePortable(schema.Not)
	makePortable(schema.If)
	makePortable(schema.Then)
	makePortable(schema.Else)
}
