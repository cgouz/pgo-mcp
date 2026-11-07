package service

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cgouz/pgo-mcp/src/modules/query/dto"
)

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

func (s *Service) ExecuteSQL(ctx context.Context, sqlQuery string) (*dto.ExecuteSQLOutput, error) {

	if strings.TrimSpace(sqlQuery) == "" {
		return nil, fmt.Errorf("sql query is required")
	}

	ctx, cancel := context.WithTimeout(ctx, s.options.Timeout)

	defer cancel()

	tx, err := s.pool.
		BeginTx(
			ctx,
			pgx.TxOptions{
				AccessMode: pgx.ReadOnly,
			},
		)
	{
		if err != nil {
			return nil, fmt.Errorf("begin transaction: %w", err)
		}

		defer tx.Rollback(context.Background())
	}

	if _, err := tx.
		Exec(
			ctx,
			"SELECT set_config('statement_timeout', $1, true)",
			s.options.Timeout.String(),
		); err != nil {
		return nil, fmt.Errorf("set timeout: %w", err)
	}

	rows, err := tx.Query(ctx, sqlQuery)
	{
		if err != nil {
			return nil, fmt.Errorf("execute sql: %w", err)
		}

		defer rows.Close()
	}

	fields := rows.FieldDescriptions()

	result := &dto.ExecuteSQLOutput{
		Columns: make([]string, 0, len(fields)),
		Rows:    make([][]any, 0),
	}

	for _, field := range fields {
		result.Columns = append(result.Columns, field.Name)
	}

	size := 0

	for rows.Next() {

		if result.RowCount >= s.options.MaxRows {
			result.Truncated = true
			break
		}

		values, err := rows.Values()
		{
			if err != nil {
				return nil, fmt.Errorf("read row: %w", err)
			}
		}

		row := make([]any, len(values))

		for i, value := range values {
			row[i] = normalize(value)
		}

		data, err := json.Marshal(row)
		{
			if err != nil {
				return nil, fmt.Errorf("encode row: %w", err)
			}
		}

		if size+len(data) > s.options.MaxBytes {
			result.Truncated = true
			break
		}

		size += len(data)

		result.Rows = append(result.Rows, row)
		result.RowCount++
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read results: %w", err)
	}

	return result, nil
}

func normalize(value any) any {
	switch v := value.(type) {
	case nil:
		return nil
	case []byte:
		return "\\x" + hex.EncodeToString(v)
	case time.Time:
		return v.Format(time.RFC3339Nano)
	default:
		return fmt.Sprint(v)
	}
}
