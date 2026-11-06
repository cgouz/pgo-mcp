package dto

type AnalyzeQueryIndexesInput struct {
	SQL string `json:"sql" jsonschema:"SQL query to analyze for index usage"`
}

type IndexScan struct {
	NodeType      string  `json:"node_type"`
	Schema        string  `json:"schema,omitempty"`
	Relation      string  `json:"relation,omitempty"`
	IndexName     string  `json:"index_name,omitempty"`
	Condition     string  `json:"condition,omitempty"`
	Filter        string  `json:"filter,omitempty"`
	EstimatedRows float64 `json:"estimated_rows"`
	TotalCost     float64 `json:"total_cost"`
}

type IndexInfo struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
	Unique     bool   `json:"unique"`
	Valid      bool   `json:"valid"`
	Ready      bool   `json:"ready"`
}

type TableIndexInfo struct {
	Schema        string      `json:"schema"`
	Table         string      `json:"table"`
	SizeBytes     int64       `json:"size_bytes"`
	EstimatedRows float64     `json:"estimated_rows"`
	Indexes       []IndexInfo `json:"indexes"`
}

type IndexFinding struct {
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type AnalyzeQueryIndexesOutput struct {
	EstimatedCost float64          `json:"estimated_cost"`
	Scans         []IndexScan      `json:"scans"`
	Tables        []TableIndexInfo `json:"tables"`
	Findings      []IndexFinding   `json:"findings"`
}
