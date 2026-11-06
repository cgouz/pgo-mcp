package dto

type GetObjectDetailsInput struct {
	SchemaName string `json:"schema_name"`
	ObjectName string `json:"object_name"`
	ObjectType string `json:"object_type,omitempty"`
}

type Column struct {
	Name       string  `json:"name"`
	DataType   string  `json:"data_type"`
	Nullable   bool    `json:"nullable"`
	Default    *string `json:"default,omitempty"`
	PrimaryKey bool    `json:"primary_key"`
}

type Constraint struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Definition string `json:"definition"`
}

type Index struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
}

type GetObjectDetailsOutput struct {
	SchemaName  string       `json:"schema_name"`
	ObjectName  string       `json:"object_name"`
	ObjectType  string       `json:"object_type"`
	Columns     []Column     `json:"columns,omitempty"`
	Constraints []Constraint `json:"constraints,omitempty"`
	Indexes     []Index      `json:"indexes,omitempty"`
	Definition  string       `json:"definition,omitempty"`
	Metadata    any          `json:"metadata,omitempty"`
}
