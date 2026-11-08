package dto

import "time"

type GetTopQueriesInput struct {
	Database string `json:"database,omitempty" jsonschema:"Target PostgreSQL database name"`
	SortBy   string `json:"sort_by,omitempty" jsonschema:"Sort by total_time, mean_time, calls or rows"`
	Limit    int    `json:"limit,omitempty" jsonschema:"Maximum queries to return, from 1 to 50"`
}

type TopQuery struct {
	QueryID     string     `json:"query_id"`
	Query       string     `json:"query"`
	Calls       int64      `json:"calls"`
	TotalExecMs float64    `json:"total_exec_ms"`
	MeanExecMs  float64    `json:"mean_exec_ms"`
	MinExecMs   float64    `json:"min_exec_ms"`
	MaxExecMs   float64    `json:"max_exec_ms"`
	Rows        int64      `json:"rows"`
	SharedHits  int64      `json:"shared_hits"`
	SharedReads int64      `json:"shared_reads"`
	StatsSince  *time.Time `json:"stats_since,omitempty"`
}

type GetTopQueriesOutput struct {
	Database string     `json:"database"`
	SortBy   string     `json:"sort_by"`
	Count    int        `json:"count"`
	Queries  []TopQuery `json:"queries"`
}
