package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	schemahandler "github.com/cgouz/pgo-mcp/src/modules/schema/handler"
	schemaservice "github.com/cgouz/pgo-mcp/src/modules/schema/service"

	objecthandler "github.com/cgouz/pgo-mcp/src/modules/object/handler"
	objectservice "github.com/cgouz/pgo-mcp/src/modules/object/service"

	queryhandler "github.com/cgouz/pgo-mcp/src/modules/query/handler"
	queryservice "github.com/cgouz/pgo-mcp/src/modules/query/service"
)

type Server struct {
	http *http.Server
}

func New(host, port string, pool *pgxpool.Pool) *Server {

	schemaService := schemaservice.New(pool)
	schemaHandler := schemahandler.New(schemaService)

	objectService := objectservice.New(pool)
	objectHandler := objecthandler.New(objectService)

	queryService := queryservice.New(pool)
	queryHandler := queryhandler.New(queryService)

	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "pgo-mcp",
			Version: "0.1.0",
		},
		nil,
	)

	{
		mcp.AddTool(
			mcpServer,
			&mcp.Tool{
				Name:        "list_schemas",
				Description: "List all schemas in the connected PostgreSQL database",
			},
			schemaHandler.ListSchemas,
		)

		mcp.AddTool(
			mcpServer,
			&mcp.Tool{
				Name:        "list_objects",
				Description: "List all objects in the connected PostgreSQL database",
			},
			objectHandler.ListObjects,
		)

		mcp.AddTool(
			mcpServer,
			&mcp.Tool{
				Name:        "get_object_details",
				Description: "Get details of a specific object in the connected PostgreSQL database",
			},
			objectHandler.GetObjectDetails,
		)

		mcp.AddTool(
			mcpServer,
			&mcp.Tool{
				Name:        "execute_sql",
				Description: "Execute a read-only SQL query against PostgreSQL",
			},
			queryHandler.ExecuteSQL,
		)

		mcp.AddTool(
			mcpServer,
			&mcp.Tool{
				Name:        "explain_query",
				Description: "Explain a PostgreSQL query execution plan, optionally collecting actual execution statistics",
			},
			queryHandler.ExplainQuery,
		)

		mcp.AddTool(
			mcpServer,
			&mcp.Tool{
				Name: "analyze_query_indexes",
				Description: "Analyze PostgreSQL query plans, " +
					"existing indexes and potential index issues",
			},
			queryHandler.AnalyzeQueryIndexes,
		)

	}

	handler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server {
			return mcpServer
		},
		nil,
	)

	mux := http.NewServeMux()

	mux.Handle("/mcp", handler)

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
