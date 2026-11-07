package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/cgouz/pgo-mcp/src/modules/query/dto"
)

const (
	defaultTopLimit = 10
	maxTopLimit     = 50
	maxQueryLength  = 2048
)

var topSortColumns = map[string]string{
	"total_time": "total_exec_time",
	"mean_time":  "mean_exec_time",
	"calls":      "calls",
	"rows":       "rows",
}

func (s *Service) GetTopQueries(ctx context.Context, input dto.GetTopQueriesInput) (*dto.GetTopQueriesOutput, error) {

	sortBy := strings.TrimSpace(input.SortBy)
	if sortBy == "" {
		sortBy = "total_time"
	}

	sortColumn, ok := topSortColumns[sortBy]
	if !ok {
		return nil, fmt.Errorf("unsupported sort_by: %s", sortBy)
	}

	limit := input.Limit
	if limit == 0 {
		limit = defaultTopLimit
	}

	if limit < 1 || limit > maxTopLimit {
		return nil, fmt.Errorf("limit must be between 1 and 50")
	}

	ctx, cancel := context.WithTimeout(ctx, s.options.Timeout)

	defer cancel()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	if _, err := tx.
		Exec(ctx,
			"SELECT set_config('statement_timeout', $1, true)",
			s.options.Timeout.String(),
		); err != nil {
		return nil, fmt.Errorf("set statement timeout: %w", err)
	}

	var schema string

	err = tx.
		QueryRow(ctx,
			`SELECT 
				n.nspname
			FROM 
				pg_catalog.pg_extension e
			JOIN 
				pg_catalog.pg_namespace n ON n.oid = e.extnamespace
			WHERE 
				e.extname = 'pg_stat_statements'`).
		Scan(&schema)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil,
			fmt.Errorf("pg_stat_statements extension is not installed")
	}
	if err != nil {
		return nil, fmt.Errorf("find extension: %w", err)
	}

	view := pgx.Identifier{
		schema,
		"pg_stat_statements",
	}.Sanitize()

	sql := fmt.Sprintf(`
		SELECT
			s.queryid::text,
			LEFT(s.query, $2),
			s.calls,
			s.total_exec_time,
			s.mean_exec_time,
			s.min_exec_time,
			s.max_exec_time,
			s.rows,
			s.shared_blks_hit,
			s.shared_blks_read,
			s.stats_since
		FROM %s AS s
		WHERE s.dbid = (
			SELECT oid
			FROM pg_catalog.pg_database
			WHERE datname = current_database()
		)
		  AND s.queryid IS NOT NULL
		  AND s.query IS NOT NULL
		  AND s.calls > 0
		  AND (
			s.userid = (
				SELECT oid
				FROM pg_catalog.pg_roles
				WHERE rolname = current_user
			)
			OR pg_catalog.pg_has_role(
				current_user,
				'pg_read_all_stats',
				'MEMBER'
			)
		  )
		ORDER BY s.%s DESC NULLS LAST, s.queryid
		LIMIT $1
	`, view, sortColumn)

	rows, err := tx.Query(ctx, sql, limit, maxQueryLength)
	if err != nil {
		return nil, fmt.Errorf("load top queries: %w", err)
	}

	defer rows.Close()

	result := &dto.GetTopQueriesOutput{
		SortBy:  sortBy,
		Queries: []dto.TopQuery{},
	}

	for rows.Next() {

		var q dto.TopQuery

		if err := rows.Scan(
			&q.QueryID,
			&q.Query,
			&q.Calls,
			&q.TotalExecMs,
			&q.MeanExecMs,
			&q.MinExecMs,
			&q.MaxExecMs,
			&q.Rows,
			&q.SharedHits,
			&q.SharedReads,
			&q.StatsSince,
		); err != nil {
			return nil, fmt.Errorf("scan top query: %w", err)
		}

		result.Queries = append(result.Queries, q)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read top queries: %w", err)
	}

	result.Count = len(result.Queries)

	if err := tx.QueryRow(ctx,
		"SELECT current_database()",
	).Scan(&result.Database); err != nil {
		return nil, fmt.Errorf("get database name: %w", err)
	}

	return result, nil
}
