package dto

type ExecuteSQLInput struct {
	Database string `json:"database,omitempty" jsonschema:"Target PostgreSQL database name"`
	SQL      string `json:"sql" jsonschema:"SQL query to execute"`
}

type ExecuteSQLOutput struct {
	Columns   []string    `json:"columns"`
	Rows      [][]*string `json:"rows"`
	RowCount  int         `json:"row_count"`
	Truncated bool        `json:"truncated"`
}
