package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/cgouz/pgo-mcp/src/modules/query/dto"
)

const largeTableBytes int64 = 16 << 20

func (s *Service) AnalyzeQueryIndexes(ctx context.Context, input dto.AnalyzeQueryIndexesInput) (*dto.AnalyzeQueryIndexesOutput, error) {

	explained, err := s.ExplainQuery(
		ctx, dto.ExplainQueryInput{
			SQL:     input.SQL,
			Analyze: false,
		},
	)
	if err != nil {
		return nil, err
	}

	result := &dto.AnalyzeQueryIndexesOutput{
		Scans:    []dto.IndexScan{},
		Tables:   []dto.TableIndexInfo{},
		Findings: []dto.IndexFinding{},
	}

	root, ok := explained.Plan[0]["Plan"].(map[string]any)
	{
		if !ok {
			return result, fmt.Errorf("invalid execution plan")
		}
	}

	result.EstimatedCost = number(root["Total Cost"])

	relations := make(map[string][2]string)

	collectScans(root, result, relations)

	for _, relation := range relations {
		table, err := s.loadTableIndexes(
			ctx,
			relation[0],
			relation[1],
		)
		if err != nil {
			return result, err
		}

		result.Tables = append(result.Tables, table)
	}

	for _, scan := range result.Scans {
		switch scan.NodeType {
		case "Index Scan", "Index Only Scan",
			"Bitmap Index Scan":

			result.Findings = append(result.Findings,
				dto.IndexFinding{
					Severity: "info",
					Message: fmt.Sprintf(
						"Planner uses index %s (%s)",
						scan.IndexName,
						scan.NodeType,
					),
				},
			)

		case "Seq Scan":
			table := findTable(
				result.Tables,
				scan.Schema,
				scan.Relation,
			)

			if table == nil || scan.Filter == "" {
				continue
			}

			if table.SizeBytes >= largeTableBytes {
				result.Findings = append(result.Findings,
					dto.IndexFinding{
						Severity: "warning",
						Message: fmt.Sprintf(
							"Filtered sequential scan on %s.%s "+
								"(table size: %d bytes). "+
								"Review filter selectivity, statistics "+
								"and existing indexes.",
							scan.Schema,
							scan.Relation,
							table.SizeBytes,
						),
					},
				)
			}
		}
	}

	if len(result.Findings) == 0 {
		result.Findings = append(result.Findings,
			dto.IndexFinding{
				Severity: "info",
				Message:  "No obvious index-related warning detected",
			},
		)
	}

	return result, nil
}

func collectScans(node map[string]any, result *dto.AnalyzeQueryIndexesOutput, relations map[string][2]string) {

	nodeType, _ := node["Node Type"].(string)
	schema, _ := node["Schema"].(string)
	relation, _ := node["Relation Name"].(string)
	indexName, _ := node["Index Name"].(string)
	condition, _ := node["Index Cond"].(string)
	filter, _ := node["Filter"].(string)

	if relation != "" && schema != "" {
		key := schema + "\x00" + relation
		relations[key] = [2]string{schema, relation}
	}

	if nodeType == "Seq Scan" ||
		nodeType == "Index Scan" ||
		nodeType == "Index Only Scan" ||
		nodeType == "Bitmap Index Scan" ||
		nodeType == "Bitmap Heap Scan" {

		result.Scans = append(result.Scans, dto.IndexScan{
			NodeType:      nodeType,
			Schema:        schema,
			Relation:      relation,
			IndexName:     indexName,
			Condition:     condition,
			Filter:        filter,
			EstimatedRows: number(node["Plan Rows"]),
			TotalCost:     number(node["Total Cost"]),
		})
	}

	children, ok := node["Plans"].([]any)
	{
		if !ok {
			return
		}
	}

	for _, child := range children {

		plan, ok := child.(map[string]any)

		if ok {
			collectScans(plan, result, relations)
		}
	}
}

func (s *Service) loadTableIndexes(ctx context.Context, schema string, name string) (dto.TableIndexInfo, error) {

	table := dto.TableIndexInfo{
		Schema:  schema,
		Table:   name,
		Indexes: []dto.IndexInfo{},
	}

	var oid uint32

	err := s.pool.
		QueryRow(ctx,
			`SELECT
				c.oid,
				pg_catalog.pg_relation_size(c.oid),
				c.reltuples::float8
			FROM 
				pg_catalog.pg_class AS c
			JOIN 
				pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace
			WHERE 
				n.nspname = $1
				AND c.relname = $2
				AND c.relkind IN ('r', 'p', 'm', 'f')
				AND has_schema_privilege(n.oid, 'USAGE')
				AND has_table_privilege(c.oid, 'SELECT')`,
			schema, name).
		Scan(
			&oid,
			&table.SizeBytes,
			&table.EstimatedRows,
		)

	if errors.Is(err, pgx.ErrNoRows) {
		return table, fmt.Errorf(
			"table not found or access denied: %s.%s",
			schema, name,
		)
	}
	if err != nil {
		return table, fmt.Errorf("load table metadata: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			i.relname,
			pg_catalog.pg_get_indexdef(i.oid),
			x.indisunique,
			x.indisvalid,
			x.indisready
		FROM 
			pg_catalog.pg_index AS x
		JOIN 
			pg_catalog.pg_class AS i ON i.oid = x.indexrelid
		WHERE 
			x.indrelid = $1::oid
		ORDER BY 
			i.relname`, oid,
	)

	if err != nil {
		return table, fmt.Errorf("load indexes: %w", err)
	}

	defer rows.Close()

	for rows.Next() {

		var index dto.IndexInfo

		if err := rows.
			Scan(
				&index.Name,
				&index.Definition,
				&index.Unique,
				&index.Valid,
				&index.Ready,
			); err != nil {

			return table, fmt.Errorf("scan index: %w", err)
		}

		table.Indexes = append(table.Indexes, index)
	}

	if err := rows.Err(); err != nil {
		return table, fmt.Errorf("read indexes: %w", err)
	}

	return table, nil
}

func findTable(tables []dto.TableIndexInfo, schema string, name string) *dto.TableIndexInfo {

	for i := range tables {
		if tables[i].Schema == schema &&
			tables[i].Table == name {
			return &tables[i]
		}
	}

	return nil
}

func number(value any) float64 {

	v, ok := value.(float64)
	if ok {
		return v
	}

	return 0
}
