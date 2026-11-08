package dto

type ListSchemasInput struct {
	Database string `json:"database,omitempty" jsonschema:"Target PostgreSQL database name"`
}

type ListSchemasOutput struct {
	Schemas []string `json:"schemas"`
}
