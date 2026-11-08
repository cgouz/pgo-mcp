package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/cgouz/pgo-mcp/src/modules/query/dto"
)

func (s *Service) ExplainQuery(ctx context.Context, input dto.ExplainQueryInput) (*dto.ExplainQueryOutput, error) {

	sql := strings.TrimSpace(input.SQL)
	{
		if sql == "" {
			return nil, fmt.Errorf("sql is required")
		}

		if len(sql) > maxSQLBytes {
			return nil, fmt.Errorf("sql exceeds size limit")
		}
	}

	ctx, cancel := context.WithTimeout(ctx, s.options.Timeout)

	defer cancel()

	tx, err := s.pool.
		BeginTx(
			ctx, pgx.TxOptions{
				AccessMode: pgx.ReadOnly,
			},
		)
	{
		if err != nil {
			return nil, fmt.Errorf("begin transaction: %w", err)
		}

		defer tx.Rollback(context.Background())
	}

	if _, err = tx.Exec(ctx,
		"SELECT set_config('statement_timeout', $1, true)",
		s.options.Timeout.String(),
	); err != nil {
		return nil, fmt.Errorf("set timeout: %w", err)
	}

	options := "FORMAT JSON, VERBOSE TRUE"

	if input.Analyze {
		options += ", ANALYZE TRUE, BUFFERS TRUE"
	}

	query := "EXPLAIN (" + options + ") " + sql

	var raw []byte
	{
		if err := tx.
			QueryRow(ctx, query).
			Scan(&raw); err != nil {
			return nil, fmt.Errorf("explain query: %w", err)
		}

		if len(raw) > s.options.MaxBytes {
			return nil, fmt.Errorf("execution plan exceeds size limit")
		}
	}

	var plan []map[string]any
	{
		if err := json.Unmarshal(raw, &plan); err != nil {
			return nil, fmt.Errorf("decode execution plan: %w", err)
		}

		if len(plan) == 0 {
			return nil, fmt.Errorf("empty execution plan")
		}
	}

	return &dto.ExplainQueryOutput{
		Analyze: input.Analyze,
		Plan:    plan,
	}, nil
}
