package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cgouz/pgo-mcp/src/modules/diagnostics/dto"
)

const (
	tableLimit = 10
	largeTable = 16 << 20
)

type Options struct {
	Timeout time.Duration
}

type Service struct {
	pool    *pgxpool.Pool
	options Options
}

func New(pool *pgxpool.Pool, opitons Options) *Service {
	return &Service{
		pool:    pool,
		options: opitons,
	}
}

func (s *Service) AnalyzeDBHealth(ctx context.Context) (*dto.AnalyzeDBHealthOutput, error) {

	ctx, cancel := context.WithTimeout(ctx, s.options.Timeout)
	defer cancel()

	tx, err := s.pool.BeginTx(
		ctx, pgx.TxOptions{
			AccessMode: pgx.ReadOnly,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	defer tx.Rollback(context.Background())

	if _, err := tx.
		Exec(ctx, "SET LOCAL statement_timeout = '10s'"); err != nil {
		return nil, fmt.Errorf("set timeout: %w", err)
	}

	result := &dto.AnalyzeDBHealthOutput{
		CollectedAt: time.Now().UTC(),
		Tables:      []dto.TableHealth{},
		Findings:    []dto.HealthFinding{},
	}

	if err := loadDatabaseStats(ctx, tx, result); err != nil {
		return nil, err
	}

	if err := loadConnections(ctx, tx, result); err != nil {
		return nil, err
	}

	if err := loadTables(ctx, tx, result); err != nil {
		return nil, err
	}

	analyzeFindings(result)

	return result, nil
}

func loadDatabaseStats(ctx context.Context, tx pgx.Tx, result *dto.AnalyzeDBHealthOutput) error {

	if err := tx.
		QueryRow(ctx,
			`SELECT
				d.datname::text,
				pg_catalog.pg_database_size(d.datid),
				d.xact_commit,
				d.xact_rollback,
				d.blks_hit,
				d.blks_read,
				d.deadlocks,
				d.temp_files,
				d.temp_bytes,
				d.stats_reset
			FROM 
				pg_catalog.pg_stat_database d
			WHERE 
				d.datname = current_database()`).
		Scan(
			&result.Database,
			&result.SizeBytes,
			&result.Transactions.Commits,
			&result.Transactions.Rollbacks,
			&result.Cache.BlocksHit,
			&result.Cache.BlocksRead,
			&result.Deadlocks,
			&result.TempFiles,
			&result.TempBytes,
			&result.StatsReset,
		); err != nil {
		return fmt.Errorf("load database statistics: %w", err)
	}

	txTotal := result.Transactions.Commits + result.Transactions.Rollbacks

	if txTotal > 0 {
		ratio := float64(result.Transactions.Rollbacks) / float64(txTotal) * 100
		result.Transactions.RollbackRatio = &ratio
	}

	blockTotal := result.Cache.BlocksHit + result.Cache.BlocksRead

	if blockTotal > 0 {
		ratio := float64(result.Cache.BlocksHit) / float64(blockTotal) * 100
		result.Cache.HitRatio = &ratio
	}

	return nil
}

func loadConnections(ctx context.Context, tx pgx.Tx, result *dto.AnalyzeDBHealthOutput) error {

	if err := tx.
		QueryRow(ctx,
			`SELECT
				COUNT(*) FILTER (
					WHERE backend_type = 'client backend'
					AND datname = current_database()
					AND state = 'active'
				),
				COUNT(*) FILTER (
					WHERE backend_type = 'client backend'
					AND datname = current_database()
					AND state = 'idle'
				),
				COUNT(*) FILTER (
					WHERE backend_type = 'client backend'
					AND datname = current_database()
					AND (
						state IS NULL OR
						state NOT IN ('active', 'idle')
					)
				),
				COUNT(*) FILTER (
					WHERE backend_type = 'client backend'
				),
				current_setting('max_connections')::int
			FROM pg_catalog.pg_stat_activity
			`).
		Scan(
			&result.Connections.DatabaseActive,
			&result.Connections.DatabaseIdle,
			&result.Connections.DatabaseOther,
			&result.Connections.ServerTotal,
			&result.Connections.MaxConnections,
		); err != nil {
		return fmt.Errorf("load connection statistics: %w", err)
	}

	if result.Connections.MaxConnections > 0 {
		result.Connections.UsagePercent = float64(result.Connections.ServerTotal) / float64(result.Connections.MaxConnections) * 100
	}

	return nil
}

func loadTables(ctx context.Context, tx pgx.Tx, result *dto.AnalyzeDBHealthOutput) error {

	rows, err := tx.
		Query(ctx,
			`SELECT
				s.schemaname::text,
				s.relname::text,
				pg_catalog.pg_total_relation_size(s.relid),
				s.n_live_tup,
				s.n_dead_tup,
				s.last_autovacuum,
				s.last_autoanalyze
			FROM 
				pg_catalog.pg_stat_user_tables s
			JOIN 
				pg_catalog.pg_namespace n ON n.nspname = s.schemaname
			WHERE 
				has_schema_privilege(n.oid, 'USAGE')
				AND has_table_privilege(s.relid, 'SELECT')
			ORDER BY
				s.n_dead_tup DESC,
				s.relid
			LIMIT $1`, tableLimit,
		)

	if err != nil {
		return fmt.Errorf("load table statistics: %w", err)
	}

	defer rows.Close()

	for rows.Next() {

		var table dto.TableHealth

		if err := rows.Scan(
			&table.Schema,
			&table.Table,
			&table.SizeBytes,
			&table.LiveTuples,
			&table.DeadTuples,
			&table.LastAutovacuum,
			&table.LastAutoanalyze,
		); err != nil {
			return fmt.Errorf("scan table statistics: %w", err)
		}

		result.Tables = append(result.Tables, table)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("read table statistics: %w", err)
	}

	return nil
}

func analyzeFindings(result *dto.AnalyzeDBHealthOutput) {

	add := func(severity, category, message string) {

		result.Findings = append(

			result.Findings,

			dto.HealthFinding{
				Severity: severity,
				Category: category,
				Message:  message,
			},
		)
	}

	if result.Connections.UsagePercent >= 80 {

		add(
			"warning",
			"connections",
			fmt.Sprintf("Server connections are at %.1f%% of max_connections", result.Connections.UsagePercent),
		)
	}

	if result.Transactions.RollbackRatio != nil && *result.Transactions.RollbackRatio >= 20 {

		add(
			"info",
			"transactions",
			"Cumulative rollback ratio is high; investigate application errors and transaction patterns",
		)
	}

	if result.Cache.HitRatio != nil && *result.Cache.HitRatio < 90 && result.Cache.BlocksHit+result.Cache.BlocksRead >= 10000 {

		add(
			"info",
			"cache",
			"Database-wide buffer cache hit ratio is below 90%; review workload and I/O before tuning",
		)
	}

	if result.Deadlocks > 0 {

		add(
			"info",
			"deadlocks",
			fmt.Sprintf("%d deadlocks recorded since the statistics reset", result.Deadlocks))
	}

	for _, table := range result.Tables {

		total := table.LiveTuples + table.DeadTuples
		{
			if total <= 0 {
				continue
			}
		}

		deadRatio := float64(table.DeadTuples) / float64(total)

		if table.SizeBytes >= largeTable && table.DeadTuples >= 1000 && deadRatio >= 0.20 {

			add(
				"warning",
				"vacuum",
				fmt.Sprintf("Table %s.%s has approximately %.1f%% "+"dead tuples; review autovacuum activity",
					table.Schema,
					table.Table,
					deadRatio*100,
				),
			)
		}
	}

	if len(result.Findings) == 0 {
		add(
			"info",
			"general",
			"No immediate issue detected by the configured health checks",
		)
	}
}
