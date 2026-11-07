package server

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"

	diagnosticsdto "github.com/cgouz/pgo-mcp/src/modules/diagnostics/dto"
	indexdto "github.com/cgouz/pgo-mcp/src/modules/index/dto"
	objectdto "github.com/cgouz/pgo-mcp/src/modules/object/dto"
	querydto "github.com/cgouz/pgo-mcp/src/modules/query/dto"
	schemadto "github.com/cgouz/pgo-mcp/src/modules/schema/dto"
)

func TestToolInputSchemasArePortable(t *testing.T) {
	schemas := map[string]*jsonschema.Schema{
		"list_schemas":             inputSchema[schemadto.ListSchemasInput](t),
		"list_objects":             inputSchema[objectdto.ListObjectsInput](t),
		"get_object_details":       inputSchema[objectdto.GetObjectDetailsInput](t),
		"execute_sql":              inputSchema[querydto.ExecuteSQLInput](t),
		"explain_query":            inputSchema[querydto.ExplainQueryInput](t),
		"analyze_query_indexes":    inputSchema[querydto.AnalyzeQueryIndexesInput](t),
		"get_top_queries":          inputSchema[querydto.GetTopQueriesInput](t),
		"analyze_db_health":        inputSchema[diagnosticsdto.AnalyzeDBHealthInput](t),
		"analyze_workload_indexes": inputSchema[indexdto.AnalyzeWorkloadInput](t),
	}

	for name, schema := range schemas {
		t.Run(name, func(t *testing.T) {
			if schema.Type != "object" {
				t.Fatalf("root schema type = %q, want object", schema.Type)
			}

			data, err := json.Marshal(schema)
			if err != nil {
				t.Fatal(err)
			}

			var document map[string]any
			if err := json.Unmarshal(data, &document); err != nil {
				t.Fatal(err)
			}

			checkPortableSchema(t, document, name)
		})
	}
}

func TestToolOutputSchemasArePortable(t *testing.T) {
	schemas := toolOutputSchemas()

	if len(schemas) != 9 {
		t.Fatalf("got %d tool output schemas, want 9", len(schemas))
	}

	for name, schema := range schemas {
		t.Run(name, func(t *testing.T) {
			if schema.Type != "object" {
				t.Fatalf("root schema type = %q, want object", schema.Type)
			}

			data, err := json.Marshal(schema)
			if err != nil {
				t.Fatal(err)
			}

			var document map[string]any
			if err := json.Unmarshal(data, &document); err != nil {
				t.Fatal(err)
			}

			checkPortableSchema(t, document, name)
		})
	}
}

func TestRepresentativeOutputsMatchSchemas(t *testing.T) {
	value := "value"
	description := (*string)(nil)

	tests := map[string]any{
		"execute_sql": querydto.ExecuteSQLOutput{
			Columns:  []string{"first", "second"},
			Rows:     [][]*string{{&value, nil}},
			RowCount: 1,
		},
		"get_object_details": objectdto.GetObjectDetailsOutput{
			SchemaName: "public",
			ObjectName: "extension",
			ObjectType: "extension",
			Metadata: &objectdto.Metadata{
				Version:     &value,
				Description: &description,
			},
		},
	}

	schemas := toolOutputSchemas()
	for name, output := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := json.Marshal(output)
			if err != nil {
				t.Fatal(err)
			}

			var instance any
			if err := json.Unmarshal(data, &instance); err != nil {
				t.Fatal(err)
			}

			resolved, err := schemas[name].Resolve(nil)
			if err != nil {
				t.Fatal(err)
			}
			if err := resolved.Validate(instance); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func checkPortableSchema(t *testing.T, schema map[string]any, path string) {
	t.Helper()

	if _, ok := schema["type"].([]any); ok {
		t.Errorf("%s.type is an array", path)
	}

	for _, field := range []string{"properties", "$defs", "definitions", "patternProperties"} {
		children, _ := schema[field].(map[string]any)
		for name, child := range children {
			checkSchemaChild(t, child, fmt.Sprintf("%s.%s.%s", path, field, name))
		}
	}

	for _, field := range []string{"items", "additionalItems", "contains", "propertyNames", "unevaluatedItems", "unevaluatedProperties", "not", "if", "then", "else"} {
		if child, ok := schema[field]; ok {
			checkSchemaChild(t, child, path+"."+field)
		}
	}

	for _, field := range []string{"prefixItems", "allOf", "anyOf", "oneOf"} {
		children, _ := schema[field].([]any)
		for i, child := range children {
			checkSchemaChild(t, child, fmt.Sprintf("%s.%s[%d]", path, field, i))
		}
	}
}

func checkSchemaChild(t *testing.T, child any, path string) {
	t.Helper()

	schema, ok := child.(map[string]any)
	if !ok {
		t.Errorf("%s is a bare boolean schema", path)
		return
	}

	checkPortableSchema(t, schema, path)
}

func inputSchema[T any](t *testing.T) *jsonschema.Schema {
	t.Helper()

	schema, err := jsonschema.For[T](nil)
	if err != nil {
		t.Fatal(err)
	}

	return schema
}
