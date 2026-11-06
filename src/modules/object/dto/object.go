package dto

type ListObjectsInput struct {
	SchemaName string `json:"schema_name" jsonschema:"PostgreSQL schema name"`
	ObjectType string `json:"object_type,omitempty" jsonschema:"Object type: table, view, sequence, or extension"`
}

type ListObjectsOutput struct {
	SchemaName string   `json:"schema_name"`
	ObjectType string   `json:"object_type"`
	Objects    []string `json:"objects"`
}
