package dto

import "time"

type AnalyzeDBHealthInput struct {
	Database string `json:"database,omitempty" jsonschema:"Target PostgreSQL database name"`
}

type Connections struct {
	DatabaseActive int     `json:"database_active"`
	DatabaseIdle   int     `json:"database_idle"`
	DatabaseOther  int     `json:"database_other"`
	ServerTotal    int     `json:"server_total"`
	MaxConnections int     `json:"max_connections"`
	UsagePercent   float64 `json:"usage_percent"`
}

type Transactions struct {
	Commits       int64    `json:"commits"`
	Rollbacks     int64    `json:"rollbacks"`
	RollbackRatio *float64 `json:"rollback_ratio,omitempty"`
}

type CacheStats struct {
	BlocksHit  int64    `json:"blocks_hit"`
	BlocksRead int64    `json:"blocks_read"`
	HitRatio   *float64 `json:"hit_ratio,omitempty"`
}

type TableHealth struct {
	Schema          string     `json:"schema"`
	Table           string     `json:"table"`
	SizeBytes       int64      `json:"size_bytes"`
	LiveTuples      int64      `json:"live_tuples"`
	DeadTuples      int64      `json:"dead_tuples"`
	LastAutovacuum  *time.Time `json:"last_autovacuum,omitempty"`
	LastAutoanalyze *time.Time `json:"last_autoanalyze,omitempty"`
}

type HealthFinding struct {
	Severity string `json:"severity"`
	Category string `json:"category"`
	Message  string `json:"message"`
}

type AnalyzeDBHealthOutput struct {
	Database     string          `json:"database"`
	SizeBytes    int64           `json:"size_bytes"`
	CollectedAt  time.Time       `json:"collected_at"`
	StatsReset   *time.Time      `json:"stats_reset,omitempty"`
	Connections  Connections     `json:"connections"`
	Transactions Transactions    `json:"transactions"`
	Cache        CacheStats      `json:"cache"`
	Deadlocks    int64           `json:"deadlocks"`
	TempFiles    int64           `json:"temp_files"`
	TempBytes    int64           `json:"temp_bytes"`
	Tables       []TableHealth   `json:"tables"`
	Findings     []HealthFinding `json:"findings"`
}
