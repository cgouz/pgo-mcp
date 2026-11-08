package dto

type GetObjectDetailsInput struct {
	Database   string `json:"database,omitempty" jsonschema:"Target PostgreSQL database name"`
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

type Metadata struct {
	DataType    *string  `json:"data_type,omitempty"`
	Start       *int64   `json:"start,omitempty"`
	Increment   *int64   `json:"increment,omitempty"`
	Minimum     *int64   `json:"min,omitempty"`
	Maximum     *int64   `json:"max,omitempty"`
	Cache       *int64   `json:"cache,omitempty"`
	Cycle       *bool    `json:"cycle,omitempty"`
	Version     *string  `json:"version,omitempty"`
	Description **string `json:"description,omitempty"`
}

type GetObjectDetailsOutput struct {
	SchemaName  string       `json:"schema_name"`
	ObjectName  string       `json:"object_name"`
	ObjectType  string       `json:"object_type"`
	Columns     []Column     `json:"columns,omitempty"`
	Constraints []Constraint `json:"constraints,omitempty"`
	Indexes     []Index      `json:"indexes,omitempty"`
	Definition  string       `json:"definition,omitempty"`
	Metadata    *Metadata    `json:"metadata,omitempty"`
}
