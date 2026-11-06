package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) ListSchemas(ctx context.Context) ([]string, error) {

	rows, err := s.pool.Query(
		ctx,
		`SELECT 
			schema_name
		FROM 
			information_schema.schemata
		ORDER BY 
			schema_name`,
	)
	{
		if err != nil {
			return nil, fmt.Errorf("list schemas: %w", err)
		}

		defer rows.Close()
	}

	schemas := make([]string, 0)

	for rows.Next() {

		var name string

		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan schema: %w", err)
		}

		schemas = append(schemas, name)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read schemas: %w", err)
	}

	return schemas, nil
}
