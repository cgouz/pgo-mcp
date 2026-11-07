package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cgouz/pgo-mcp/src/modules/index/dto"
)

const (
	defaultLimit    = 20
	maxLimit        = 50
	largeIndexBytes = 16 << 20
)

type Options struct {
	Timeout time.Duration
}

type Service struct {
	pool    *pgxpool.Pool
	options Options
}

func New(pool *pgxpool.Pool, options Options) *Service {
	return &Service{
		pool:    pool,
		options: options,
	}
}

func (s *Service) AnalyzeWorkload(
	ctx context.Context,
	input dto.AnalyzeWorkloadInput,
) (*dto.AnalyzeWorkloadOutput, error) {
	limit := input.Limit

	if limit == 0 {
		limit = defaultLimit
	}

	if limit < 1 || limit > maxLimit {
		return nil, fmt.Errorf(
			"limit must be between 1 and %d", maxLimit,
		)
	}

	if s.options.Timeout <= 0 {
		return nil, fmt.Errorf("query timeout must be positive")
	}

	ctx, cancel := context.WithTimeout(
		ctx, s.options.Timeout,
	)
	defer cancel()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(ctx,
		"SELECT set_config('statement_timeout', $1, true)",
		s.options.Timeout.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("set statement timeout: %w", err)
	}

	result := &dto.AnalyzeWorkloadOutput{
		Queries:  []dto.WorkloadQuery{},
		Indexes:  []dto.IndexUsage{},
		Findings: []dto.Finding{},
	}

	if err := tx.QueryRow(ctx,
		"SELECT current_database()",
	).Scan(&result.Database); err != nil {
		return nil, fmt.Errorf("get database: %w", err)
	}

	schema, err := findExtension(ctx, tx)
	if err != nil {
		return nil, err
	}

	if err := loadQueries(
		ctx, tx, schema, limit, result,
	); err != nil {
		return nil, err
	}

	if err := loadIndexes(
		ctx, tx, limit, result,
	); err != nil {
		return nil, err
	}

	analyzeFindings(result)

	return result, nil
}

func findExtension(
	ctx context.Context,
	tx pgx.Tx,
) (string, error) {
	var schema string

	err := tx.QueryRow(ctx, `
		SELECT n.nspname
		FROM pg_catalog.pg_extension e
		JOIN pg_catalog.pg_namespace n
			ON n.oid = e.extnamespace
		WHERE e.extname = 'pg_stat_statements'
	`).Scan(&schema)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf(
			"pg_stat_statements extension is not installed",
		)
	}
	if err != nil {
		return "", fmt.Errorf("find extension: %w", err)
	}

	return schema, nil
}

func loadQueries(
	ctx context.Context,
	tx pgx.Tx,
	schema string,
	limit int,
	result *dto.AnalyzeWorkloadOutput,
) error {
	view := pgx.Identifier{
		schema,
		"pg_stat_statements",
	}.Sanitize()

	query := fmt.Sprintf(`
		SELECT
			s.queryid::text,
			LEFT(s.query, 2048),
			s.calls,
			s.total_exec_time,
			s.mean_exec_time,
			s.rows,
			s.shared_blks_hit,
			s.shared_blks_read,
			s.temp_blks_written
		FROM %s AS s
		WHERE s.dbid = (
			SELECT oid
			FROM pg_catalog.pg_database
			WHERE datname = current_database()
		)
		  AND s.userid = (
			SELECT oid
			FROM pg_catalog.pg_roles
			WHERE rolname = current_user
		  )
		  AND s.queryid IS NOT NULL
		  AND s.query IS NOT NULL
		  AND s.calls > 0
		ORDER BY s.total_exec_time DESC, s.queryid
		LIMIT $1
	`, view)

	rows, err := tx.Query(ctx, query, limit)
	if err != nil {
		return fmt.Errorf("load workload queries: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var q dto.WorkloadQuery

		if err := rows.Scan(
			&q.QueryID,
			&q.Query,
			&q.Calls,
			&q.TotalExecMs,
			&q.MeanExecMs,
			&q.Rows,
			&q.SharedHits,
			&q.SharedReads,
			&q.TempWritten,
		); err != nil {
			return fmt.Errorf("scan workload query: %w", err)
		}

		result.Queries = append(result.Queries, q)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("read workload queries: %w", err)
	}

	return nil
}

func loadIndexes(
	ctx context.Context,
	tx pgx.Tx,
	limit int,
	result *dto.AnalyzeWorkloadOutput,
) error {
	rows, err := tx.Query(ctx, `
		SELECT
			s.schemaname::text,
			s.relname::text,
			s.indexrelname::text,
			COALESCE(s.idx_scan, 0),
			pg_catalog.pg_relation_size(s.indexrelid),
			i.indisprimary,
			i.indisunique
		FROM pg_catalog.pg_stat_user_indexes s
		JOIN pg_catalog.pg_index i
			ON i.indexrelid = s.indexrelid
		JOIN pg_catalog.pg_namespace n
			ON n.nspname = s.schemaname
		WHERE has_schema_privilege(n.oid, 'USAGE')
		  AND has_table_privilege(s.relid, 'SELECT')
		ORDER BY
			COALESCE(s.idx_scan, 0),
			pg_catalog.pg_relation_size(s.indexrelid) DESC,
			s.indexrelname
		LIMIT $1
	`, limit)
	if err != nil {
		return fmt.Errorf("load index usage: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var index dto.IndexUsage

		if err := rows.Scan(
			&index.Schema,
			&index.Table,
			&index.Index,
			&index.Scans,
			&index.SizeBytes,
			&index.Primary,
			&index.Unique,
		); err != nil {
			return fmt.Errorf("scan index usage: %w", err)
		}

		result.Indexes = append(result.Indexes, index)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("read index usage: %w", err)
	}

	return nil
}

func analyzeFindings(result *dto.AnalyzeWorkloadOutput) {
	for _, index := range result.Indexes {
		if index.Scans != 0 ||
			index.SizeBytes < largeIndexBytes ||
			index.Primary ||
			index.Unique {
			continue
		}

		result.Findings = append(
			result.Findings,
			dto.Finding{
				Severity: "warning",
				Message: fmt.Sprintf(
					"Index %s.%s has zero recorded scans "+
						"and occupies %d bytes. "+
						"Review statistics history and workload "+
						"before considering removal.",
					index.Schema,
					index.Index,
					index.SizeBytes,
				),
			},
		)
	}

	for _, q := range result.Queries {
		if q.Calls < 100 || q.MeanExecMs < 100 {
			continue
		}

		result.Findings = append(
			result.Findings,
			dto.Finding{
				Severity: "warning",
				Message: fmt.Sprintf(
					"Query %s has %d calls and "+
						"mean execution time %.2f ms. "+
						"Review its execution plan.",
					q.QueryID,
					q.Calls,
					q.MeanExecMs,
				),
			},
		)
	}

	if len(result.Findings) == 0 {
		result.Findings = append(
			result.Findings,
			dto.Finding{
				Severity: "info",
				Message:  "No obvious workload or index warning detected",
			},
		)
	}
}
