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

const (
	queryTimeout = 10 * time.Second
	maxRows      = 100
	maxBytes     = 1 << 20
)

type Service struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) ExecuteSQL(ctx context.Context, sqlQuery string) (*dto.ExecuteSQLOutput, error) {

	if strings.TrimSpace(sqlQuery) == "" {
		return nil, fmt.Errorf("sql query is required")
	}

	ctx, cancel := context.WithTimeout(ctx, queryTimeout)

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
			"SET LOCAL statement_timeout = '10s'",
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

		if result.RowCount >= maxRows {
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

		if size+len(data) > maxBytes {
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
