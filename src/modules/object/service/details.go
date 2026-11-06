package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/cgouz/pgo-mcp/src/modules/object/dto"
)

func (s *Service) GetObjectDetails(ctx context.Context, input dto.GetObjectDetailsInput) (dto.GetObjectDetailsOutput, error) {

	schema := strings.TrimSpace(input.SchemaName)
	name := strings.TrimSpace(input.ObjectName)
	kind := input.ObjectType

	if schema == "" || name == "" {
		err := fmt.Errorf("schema_name and object_name are required")
		return dto.GetObjectDetailsOutput{}, err
	}

	if kind == "" {
		kind = "table"
	}

	switch kind {
	case "table", "view", "sequence", "extension":
	default:
		err := fmt.Errorf("unsupported object type: %s", kind)
		return dto.GetObjectDetailsOutput{}, err
	}

	result := dto.GetObjectDetailsOutput{
		SchemaName:  schema,
		ObjectName:  name,
		ObjectType:  kind,
		Columns:     []dto.Column{},
		Constraints: []dto.Constraint{},
		Indexes:     []dto.Index{},
	}

	if kind == "extension" {
		err := s.extensionDetails(ctx, &result)
		return result, err
	}

	var oid uint32

	err := s.pool.
		QueryRow(ctx,
			`SELECT 
				c.oid
			FROM 
				pg_catalog.pg_class AS c
			JOIN 
				pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace
			WHERE 
				n.nspname = $1
				AND c.relname = $2
				AND (
					($3 = 'table' AND c.relkind IN ('r', 'p'))
					OR ($3 = 'view' AND c.relkind IN ('v', 'm'))
					OR ($3 = 'sequence' AND c.relkind = 'S')
				)
				AND has_schema_privilege(n.oid, 'USAGE')
				AND (
					($3 = 'sequence' AND
					has_sequence_privilege(c.oid, 'USAGE, SELECT'))
					OR ($3 <> 'sequence' AND has_table_privilege(c.oid, 'SELECT'))
				)`, schema, name, kind).
		Scan(&oid)

	if errors.Is(err, pgx.ErrNoRows) {
		return result, fmt.Errorf("object not found or access denied")
	}

	if err != nil {
		return result, fmt.Errorf("find object: %w", err)
	}

	if kind == "sequence" {
		err = s.sequenceDetails(ctx, oid, &result)
		return result, err
	}

	if err := s.loadColumns(ctx, oid, &result); err != nil {
		return result, err
	}

	if err := s.loadConstraints(ctx, oid, &result); err != nil {
		return result, err
	}

	if err := s.loadIndexes(ctx, oid, &result); err != nil {
		return result, err
	}

	if kind == "view" {
		if err = s.pool.
			QueryRow(ctx, "SELECT pg_get_viewdef($1::oid, true)", oid).
			Scan(&result.Definition); err != nil {
			return result, fmt.Errorf("view definition: %w", err)
		}
	}

	return result, nil
}

func (s *Service) loadColumns(ctx context.Context, oid uint32, result *dto.GetObjectDetailsOutput) error {

	rows, err := s.pool.
		Query(ctx, `
			SELECT
				a.attname,
				pg_catalog.format_type(a.atttypid, a.atttypmod),
				NOT a.attnotnull,
				pg_catalog.pg_get_expr(d.adbin, d.adrelid),
				EXISTS (
					SELECT 1
					FROM pg_catalog.pg_constraint co
					WHERE co.conrelid = a.attrelid
					AND co.contype = 'p'
					AND a.attnum = ANY(co.conkey)
				)
			FROM 
				pg_catalog.pg_attribute AS a
			LEFT JOIN 
				pg_catalog.pg_attrdef AS d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
			WHERE 
				a.attrelid = $1::oid
				AND a.attnum > 0
				AND NOT a.attisdropped
			ORDER BY 
				a.attnum`, oid)
	{
		if err != nil {
			return fmt.Errorf("get columns: %w", err)
		}
	}

	defer rows.Close()

	for rows.Next() {

		var col dto.Column

		err := rows.Scan(
			&col.Name,
			&col.DataType,
			&col.Nullable,
			&col.Default,
			&col.PrimaryKey,
		)
		{
			if err != nil {
				return fmt.Errorf("scan column: %w", err)
			}
		}

		result.Columns = append(result.Columns, col)
	}

	return rows.Err()
}

func (s *Service) loadConstraints(ctx context.Context, oid uint32, result *dto.GetObjectDetailsOutput) error {

	rows, err := s.pool.Query(ctx, `
		SELECT
			conname,
			CASE contype
				WHEN 'p' THEN 'PRIMARY KEY'
				WHEN 'f' THEN 'FOREIGN KEY'
				WHEN 'u' THEN 'UNIQUE'
				WHEN 'c' THEN 'CHECK'
				WHEN 'x' THEN 'EXCLUDE'
				ELSE contype::text
			END,
			pg_catalog.pg_get_constraintdef(oid, true)
		FROM 
			pg_catalog.pg_constraint
		WHERE 
			conrelid = $1::oid
		ORDER BY 
			conname`, oid)
	{
		if err != nil {
			return fmt.Errorf("get constraints: %w", err)
		}
	}

	defer rows.Close()

	for rows.Next() {

		var c dto.Constraint

		if err := rows.Scan(
			&c.Name,
			&c.Type,
			&c.Definition,
		); err != nil {
			return fmt.Errorf("scan constraint: %w", err)
		}

		result.Constraints = append(result.Constraints, c)
	}

	return rows.Err()
}

func (s *Service) loadIndexes(
	ctx context.Context,
	oid uint32,
	result *dto.GetObjectDetailsOutput,
) error {
	rows, err := s.pool.Query(ctx, `
		SELECT
			i.relname,
			pg_catalog.pg_get_indexdef(i.oid)
		FROM pg_catalog.pg_index x
		JOIN pg_catalog.pg_class i
			ON i.oid = x.indexrelid
		WHERE x.indrelid = $1::oid
		ORDER BY i.relname
	`, oid)
	if err != nil {
		return fmt.Errorf("get indexes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var index dto.Index

		if err := rows.Scan(
			&index.Name,
			&index.Definition,
		); err != nil {
			return fmt.Errorf("scan index: %w", err)
		}

		result.Indexes = append(result.Indexes, index)
	}

	return rows.Err()
}

func (s *Service) sequenceDetails(
	ctx context.Context,
	oid uint32,
	result *dto.GetObjectDetailsOutput,
) error {
	var dataType string
	var start, increment, minimum, maximum, cache int64
	var cycle bool

	err := s.pool.QueryRow(ctx, `
		SELECT
			pg_catalog.format_type(seqtypid, NULL),
			seqstart,
			seqincrement,
			seqmin,
			seqmax,
			seqcache,
			seqcycle
		FROM pg_catalog.pg_sequence
		WHERE seqrelid = $1::oid
	`, oid).Scan(
		&dataType,
		&start,
		&increment,
		&minimum,
		&maximum,
		&cache,
		&cycle,
	)
	if err != nil {
		return fmt.Errorf("get sequence: %w", err)
	}

	result.Metadata = map[string]any{
		"data_type": dataType,
		"start":     start,
		"increment": increment,
		"min":       minimum,
		"max":       maximum,
		"cache":     cache,
		"cycle":     cycle,
	}

	return nil
}

func (s *Service) extensionDetails(
	ctx context.Context,
	result *dto.GetObjectDetailsOutput,
) error {
	var version string
	var description *string

	err := s.pool.QueryRow(ctx, `
		SELECT
			e.extversion,
			d.description
		FROM pg_catalog.pg_extension e
		JOIN pg_catalog.pg_namespace n
			ON n.oid = e.extnamespace
		LEFT JOIN pg_catalog.pg_description d
			ON d.objoid = e.oid
			AND d.classoid = 'pg_catalog.pg_extension'::regclass
		WHERE n.nspname = $1
		  AND e.extname = $2
		  AND has_schema_privilege(n.oid, 'USAGE')
	`, result.SchemaName, result.ObjectName).Scan(
		&version,
		&description,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("extension not found or access denied")
	}
	if err != nil {
		return fmt.Errorf("get extension: %w", err)
	}

	result.Metadata = map[string]any{
		"version":     version,
		"description": description,
	}

	return nil
}
