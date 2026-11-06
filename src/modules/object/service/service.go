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

func (s *Service) ListObjects(ctx context.Context, schemaName string, objectType string) ([]string, error) {

	if schemaName == "" {
		return nil, fmt.Errorf("schema_name is required")
	}

	if objectType == "" {
		objectType = "table"
	}

	var query string

	switch objectType {
	case "table":

		query = `
			SELECT 
				c.relname
			FROM 
				pg_catalog.pg_class AS c
			JOIN 
				pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace
			WHERE 
				n.nspname = $1
			  	AND c.relkind IN ('r', 'p')
			  	AND has_schema_privilege(n.oid, 'USAGE')
			  	AND has_table_privilege(c.oid, 'SELECT')
			ORDER BY 
				c.relname
		`

	case "view":

		query = `
			SELECT 
				c.relname
			FROM 
				pg_catalog.pg_class AS c
			JOIN 
				pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace
			WHERE 
				n.nspname = $1
			  	AND c.relkind IN ('v', 'm')
			  	AND has_schema_privilege(n.oid, 'USAGE')
			  	AND has_table_privilege(c.oid, 'SELECT')
			ORDER BY 
				c.relname
		`

	case "sequence":

		query = `
			SELECT 
				c.relname
			FROM 
				pg_catalog.pg_class AS c
			JOIN 
				pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace
			WHERE 
				n.nspname = $1
			  	AND c.relkind = 'S'
			  	AND has_schema_privilege(n.oid, 'USAGE')
			  	AND has_sequence_privilege(c.oid, 'USAGE, SELECT')
			ORDER BY 
				c.relname
		`

	case "extension":

		query = `
			SELECT 
				e.extname
			FROM 
				pg_catalog.pg_extension AS e
			JOIN 
				pg_catalog.pg_namespace AS n ON n.oid = e.extnamespace
			WHERE 
				n.nspname = $1 
				AND has_schema_privilege(n.oid, 'USAGE')
			ORDER BY 
				e.extname
		`

	default:
		return nil, fmt.Errorf(
			"unsupported object type: %s", objectType,
		)
	}

	rows, err := s.pool.Query(ctx, query, schemaName)
	{
		if err != nil {
			return nil, fmt.Errorf("list objects: %w", err)
		}

		defer rows.Close()
	}

	objects := make([]string, 0)

	for rows.Next() {

		var name string

		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan object: %w", err)
		}

		objects = append(objects, name)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read objects: %w", err)
	}

	return objects, nil
}
