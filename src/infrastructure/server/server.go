package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/cgouz/pgo-mcp/src/infrastructure/auth"
	"github.com/cgouz/pgo-mcp/src/infrastructure/config"
	schemahandler "github.com/cgouz/pgo-mcp/src/modules/schema/handler"
	schemaservice "github.com/cgouz/pgo-mcp/src/modules/schema/service"

	objecthandler "github.com/cgouz/pgo-mcp/src/modules/object/handler"
	objectservice "github.com/cgouz/pgo-mcp/src/modules/object/service"

	queryhandler "github.com/cgouz/pgo-mcp/src/modules/query/handler"
	queryservice "github.com/cgouz/pgo-mcp/src/modules/query/service"

	diagnosticshandler "github.com/cgouz/pgo-mcp/src/modules/diagnostics/handler"
	diagnosticsservice "github.com/cgouz/pgo-mcp/src/modules/diagnostics/service"

	indexhandler "github.com/cgouz/pgo-mcp/src/modules/index/handler"
	indexservice "github.com/cgouz/pgo-mcp/src/modules/index/service"
)

type Server struct {
	http *http.Server
}

func New(host, port string, pool *pgxpool.Pool, authToken string, cfg config.Config) *Server {

	schemaService := schemaservice.New(pool)
	schemaHandler := schemahandler.New(schemaService)

	objectService := objectservice.New(pool)
	objectHandler := objecthandler.New(objectService)

	queryService := queryservice.New(
		pool,
		queryservice.Options{
			Timeout:  cfg.QueryTimeout,
			MaxRows:  cfg.QueryMaxRows,
			MaxBytes: cfg.QueryMaxBytes,
		},
	)
	queryHandler := queryhandler.New(queryService)

	diagnosticsService := diagnosticsservice.New(
		pool,
		diagnosticsservice.Options{
			Timeout: cfg.QueryTimeout,
		},
	)
	diagnosticsHandler := diagnosticshandler.New(diagnosticsService)

	indexService := indexservice.New(
		pool,
		indexservice.Options{
			Timeout: cfg.QueryTimeout,
		},
	)

	indexHandler := indexhandler.New(indexService)

	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "pgo-mcp",
			Version: "0.1.0",
		},
		nil,
	)
	outputSchemas := toolOutputSchemas()

	{
		mcp.AddTool(
			mcpServer,
			&mcp.Tool{
				Name:         "list_schemas",
				Description:  "List all schemas in the connected PostgreSQL database",
				OutputSchema: outputSchemas["list_schemas"],
			},
			schemaHandler.ListSchemas,
		)

		mcp.AddTool(
			mcpServer,
			&mcp.Tool{
				Name:         "list_objects",
				Description:  "List all objects in the connected PostgreSQL database",
				OutputSchema: outputSchemas["list_objects"],
			},
			objectHandler.ListObjects,
		)

		mcp.AddTool(
			mcpServer,
			&mcp.Tool{
				Name:         "get_object_details",
				Description:  "Get details of a specific object in the connected PostgreSQL database",
				OutputSchema: outputSchemas["get_object_details"],
			},
			objectHandler.GetObjectDetails,
		)

		mcp.AddTool(
			mcpServer,
			&mcp.Tool{
				Name:         "execute_sql",
				Description:  "Execute a read-only SQL query against PostgreSQL",
				OutputSchema: outputSchemas["execute_sql"],
			},
			queryHandler.ExecuteSQL,
		)

		mcp.AddTool(
			mcpServer,
			&mcp.Tool{
				Name:         "explain_query",
				Description:  "Explain a PostgreSQL query execution plan, optionally collecting actual execution statistics",
				OutputSchema: outputSchemas["explain_query"],
			},
			queryHandler.ExplainQuery,
		)

		mcp.AddTool(
			mcpServer,
			&mcp.Tool{
				Name: "analyze_query_indexes",
				Description: "Analyze PostgreSQL query plans, " +
					"existing indexes and potential index issues",
				OutputSchema: outputSchemas["analyze_query_indexes"],
			},
			queryHandler.AnalyzeQueryIndexes,
		)

		mcp.AddTool(
			mcpServer,
			&mcp.Tool{
				Name: "get_top_queries",
				Description: "Get PostgreSQL queries ranked by total " +
					"execution time, mean execution time, calls or rows",
				OutputSchema: outputSchemas["get_top_queries"],
			},
			queryHandler.GetTopQueries,
		)

		mcp.AddTool(
			mcpServer,
			&mcp.Tool{
				Name: "analyze_db_health",
				Description: "Analyze PostgreSQL database health, " +
					"connections, transactions, cache, deadlocks " +
					"and vacuum statistics",
				OutputSchema: outputSchemas["analyze_db_health"],
			},
			diagnosticsHandler.AnalyzeDBHealth,
		)

		mcp.AddTool(
			mcpServer,
			&mcp.Tool{
				Name: "analyze_workload_indexes",
				Description: "Analyze PostgreSQL workload statistics, " +
					"query execution time and index usage",
				OutputSchema: outputSchemas["analyze_workload_indexes"],
			},
			indexHandler.AnalyzeWorkloadIndexes,
		)
	}

	handler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server {
			return mcpServer
		},
		nil,
	)

	mux := http.NewServeMux()

	mux.Handle("/mcp", auth.Bearer(authToken, handler))

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	return &Server{
		http: &http.Server{
			Addr:              host + ":" + port,
			Handler:           mux,
			ReadHeaderTimeout: 5 * 1e9,
		},
	}
}

func (s *Server) Run() error {

	err := s.http.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

func (s *Server) Shutdown(ctx context.Context) error {

	if err := s.http.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}

	return nil
}
