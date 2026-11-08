package dto

type ExplainQueryInput struct {
	Database string `json:"database,omitempty" jsonschema:"Target PostgreSQL database name"`
	SQL      string `json:"sql" jsonschema:"SQL query to explain"`
	Analyze  bool   `json:"analyze,omitempty" jsonschema:"Execute the query and collect actual execution statistics"`
}

type ExplainQueryOutput struct {
	Analyze bool             `json:"analyze"`
	Plan    []map[string]any `json:"plan"`
}
