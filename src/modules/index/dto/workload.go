package dto

type AnalyzeWorkloadInput struct {
	Limit int `json:"limit,omitempty"`
}

type WorkloadQuery struct {
	QueryID     string  `json:"query_id"`
	Query       string  `json:"query"`
	Calls       int64   `json:"calls"`
	TotalExecMs float64 `json:"total_exec_ms"`
	MeanExecMs  float64 `json:"mean_exec_ms"`
	Rows        int64   `json:"rows"`
	SharedHits  int64   `json:"shared_hits"`
	SharedReads int64   `json:"shared_reads"`
	TempWritten int64   `json:"temp_written"`
}

type IndexUsage struct {
	Schema    string `json:"schema"`
	Table     string `json:"table"`
	Index     string `json:"index"`
	Scans     int64  `json:"scans"`
	SizeBytes int64  `json:"size_bytes"`
	Primary   bool   `json:"primary"`
	Unique    bool   `json:"unique"`
}

type Finding struct {
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type AnalyzeWorkloadOutput struct {
	Database string          `json:"database"`
	Queries  []WorkloadQuery `json:"queries"`
	Indexes  []IndexUsage    `json:"indexes"`
	Findings []Finding       `json:"findings"`
}
