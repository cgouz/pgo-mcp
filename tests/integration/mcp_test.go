package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/cgouz/pgo-mcp/src/infrastructure/database"
	queryservice "github.com/cgouz/pgo-mcp/src/modules/query/service"
)

const defaultEndpoint = "http://127.0.0.1:8002/mcp"

func endpoint() string {
	if value := os.Getenv("MCP_TEST_ENDPOINT"); value != "" {
		return value
	}
	return defaultEndpoint
}

type bearerTransport struct {
	base  http.RoundTripper
	token string
}

func (t bearerTransport) RoundTrip(
	req *http.Request,
) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header.Set(
		"Authorization",
		"Bearer "+t.token,
	)
	return t.base.RoundTrip(clone)
}

func testSession(t *testing.T) (
	context.Context,
	*mcp.ClientSession,
) {
	t.Helper()

	token := os.Getenv("MCP_AUTH_TOKEN")
	if token == "" {
		t.Fatal("MCP_AUTH_TOKEN is required")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	t.Cleanup(cancel)

	httpClient := &http.Client{
		Transport: bearerTransport{
			base:  http.DefaultTransport,
			token: token,
		},
	}

	transport := &mcp.StreamableClientTransport{
		Endpoint:   endpoint(),
		HTTPClient: httpClient,
	}

	client := mcp.NewClient(
		&mcp.Implementation{
			Name:    "pgo-mcp-integration-test",
			Version: "1.0.0",
		},
		nil,
	)

	session, err := client.Connect(
		ctx,
		transport,
		nil,
	)
	if err != nil {
		t.Fatalf("connect MCP client: %v", err)
	}

	t.Cleanup(func() {
		_ = session.Close()
	})

	return ctx, session
}

func callTool(
	t *testing.T,
	ctx context.Context,
	session *mcp.ClientSession,
	name string,
	args map[string]any,
) map[string]any {
	t.Helper()

	result, err := session.CallTool(
		ctx,
		&mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		},
	)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if result.IsError {
		t.Fatalf(
			"%s returned tool error: %s",
			name,
			toolErrorContent(result.Content),
		)
	}

	data, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatalf("%s marshal: %v", name, err)
	}

	var output map[string]any
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("%s decode: %v", name, err)
	}
	if output == nil {
		t.Fatalf("%s returned empty structured content", name)
	}

	return output
}

func toolErrorContent(content []mcp.Content) string {
	texts := make([]string, 0, len(content))

	for _, item := range content {
		if value, ok := item.(*mcp.TextContent); ok {
			texts = append(texts, value.Text)
		}
	}

	if len(texts) > 0 {
		return strings.Join(texts, "\n")
	}

	data, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return fmt.Sprintf("unreadable MCP content: %v", err)
	}

	return string(data)
}

func requireToolError(
	t *testing.T,
	ctx context.Context,
	session *mcp.ClientSession,
	name string,
	args map[string]any,
) {
	t.Helper()

	result, err := session.CallTool(
		ctx,
		&mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		},
	)

	if err == nil && (result == nil || !result.IsError) {
		t.Fatalf("%s: expected error, got success", name)
	}
}

func requireNonEmptyArray(
	t *testing.T,
	output map[string]any,
	field string,
) {
	t.Helper()

	items, ok := output[field].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf(
			"expected non-empty %s, got %v",
			field,
			output[field],
		)
	}
}

func TestAuthentication(t *testing.T) {
	for _, tc := range []struct {
		name  string
		token string
	}{
		{"missing", ""},
		{"invalid", "invalid-test-token"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(
				http.MethodGet,
				endpoint(),
				nil,
			)
			if err != nil {
				t.Fatal(err)
			}

			if tc.token != "" {
				req.Header.Set(
					"Authorization",
					"Bearer "+tc.token,
				)
			}

			client := &http.Client{
				Timeout: 5 * time.Second,
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			io.Copy(io.Discard, resp.Body)

			if resp.StatusCode != http.StatusUnauthorized {
				t.Fatalf(
					"expected 401, got %d",
					resp.StatusCode,
				)
			}
		})
	}
}

func TestToolDiscovery(t *testing.T) {
	ctx, session := testSession(t)

	result, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"list_databases",
		"list_schemas",
		"list_objects",
		"get_object_details",
		"execute_sql",
		"explain_query",
		"analyze_query_indexes",
		"analyze_workload_indexes",
		"analyze_db_health",
		"get_top_queries",
	}

	found := make(map[string]bool)

	for _, tool := range result.Tools {
		found[tool.Name] = true
	}

	for _, name := range expected {
		if !found[name] {
			t.Errorf("missing MCP tool: %s", name)
		}
	}
}

func TestDatabaseDiscoveryAndSelection(t *testing.T) {
	ctx, session := testSession(t)

	listed := callTool(
		t, ctx, session, "list_databases",
		map[string]any{},
	)
	if listed["default"] != "pgo_mcp_test" {
		t.Fatalf("unexpected default database: %v", listed["default"])
	}

	databases, ok := listed["databases"].([]any)
	if !ok || len(databases) != 2 {
		t.Fatalf("unexpected databases: %v", listed["databases"])
	}

	for _, test := range []struct {
		name     string
		database string
		expected string
	}{
		{name: "default", expected: "pgo_mcp_test"},
		{name: "explicit", database: "pgo_mcp_test_2", expected: "pgo_mcp_test_2"},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := callTool(
				t, ctx, session, "execute_sql",
				map[string]any{
					"database": test.database,
					"sql":      "SELECT current_database()",
				},
			)

			rows, ok := result["rows"].([]any)
			if !ok || len(rows) != 1 {
				t.Fatalf("unexpected rows: %v", result["rows"])
			}
			first, ok := rows[0].([]any)
			if !ok || len(first) != 1 || first[0] != test.expected {
				t.Fatalf("unexpected database result: %v", rows)
			}
		})
	}

	for _, name := range []string{
		"pgo_mcp_denied",
		"missing_database",
		"postgresql://other:secret@evil.example/database",
	} {
		requireToolError(
			t, ctx, session, "execute_sql",
			map[string]any{
				"database": name,
				"sql":      "SELECT current_database()",
			},
		)
	}

	indexes := callTool(
		t, ctx, session, "analyze_query_indexes",
		map[string]any{
			"database": "pgo_mcp_test_2",
			"sql":      "SELECT * FROM public.procurements WHERE status = 'secondary'",
		},
	)
	requireNonEmptyArray(t, indexes, "tables")

	requireToolError(
		t, ctx, session, "execute_sql",
		map[string]any{
			"database": "pgo_mcp_test_2",
			"sql":      "DELETE FROM public.procurements",
		},
	)
}

func TestDatabaseRegistry(t *testing.T) {
	ctx := context.Background()
	uri := os.Getenv("DATABASE_URI")

	t.Run("pool reuse and limit", func(t *testing.T) {
		registry, err := database.NewRegistry(
			ctx,
			uri,
			database.RegistryOptions{
				Discovery: true,
				AllowAll:  true,
				MaxPools:  2,
				MaxConns:  2,
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(registry.Close)

		first, err := registry.Pool(ctx, "pgo_mcp_test_2")
		if err != nil {
			t.Fatal(err)
		}
		second, err := registry.Pool(ctx, "pgo_mcp_test_2")
		if err != nil {
			t.Fatal(err)
		}
		if first != second {
			t.Fatal("database pool was not reused")
		}

		var wg sync.WaitGroup
		errs := make(chan error, 20)
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func(useDefault bool) {
				defer wg.Done()

				name := "pgo_mcp_test_2"
				expected := name
				if useDefault {
					name = ""
					expected = "pgo_mcp_test"
				}

				pool, err := registry.Pool(ctx, name)
				if err != nil {
					errs <- err
					return
				}

				var actual string
				if err := pool.QueryRow(ctx, "SELECT current_database()").Scan(&actual); err != nil {
					errs <- err
					return
				}
				if actual != expected {
					errs <- fmt.Errorf("expected %s, got %s", expected, actual)
				}
			}(i%2 == 0)
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			t.Error(err)
		}

		if _, err := registry.Pool(ctx, "postgres"); err == nil {
			t.Fatal("expected database pool limit error")
		}

		registry.Close()
		if err := first.Ping(ctx); err == nil {
			t.Fatal("database pool remained open after registry close")
		}
	})

	t.Run("concurrent pool creation", func(t *testing.T) {
		registry, err := database.NewRegistry(
			ctx,
			uri,
			database.RegistryOptions{
				Discovery: true,
				AllowAll:  true,
				MaxPools:  2,
				MaxConns:  2,
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		defer registry.Close()

		var wg sync.WaitGroup
		pools := make(chan *pgxpool.Pool, 10)
		errs := make(chan error, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				pool, err := registry.Pool(ctx, "pgo_mcp_test_2")
				if err != nil {
					errs <- err
					return
				}
				pools <- pool
			}()
		}

		wg.Wait()
		close(pools)
		close(errs)

		for err := range errs {
			t.Error(err)
		}

		var expected *pgxpool.Pool
		for pool := range pools {
			if expected == nil {
				expected = pool
			}
			if pool != expected {
				t.Fatal("concurrent request created a duplicate pool")
			}
		}
	})

	t.Run("default only", func(t *testing.T) {
		registry, err := database.NewRegistry(
			ctx,
			uri,
			database.RegistryOptions{
				MaxPools: 1,
				MaxConns: 2,
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		defer registry.Close()

		items, err := registry.Databases(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(items) != 1 || items[0].Name != "pgo_mcp_test" {
			t.Fatalf("unexpected default-only databases: %v", items)
		}
		if _, err := registry.Pool(ctx, "pgo_mcp_test_2"); err == nil {
			t.Fatal("default-only registry accepted another database")
		}
	})

	t.Run("wildcard discovery", func(t *testing.T) {
		registry, err := database.NewRegistry(
			ctx,
			uri,
			database.RegistryOptions{
				Discovery: true,
				AllowAll:  true,
				MaxPools:  5,
				MaxConns:  2,
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		defer registry.Close()

		items, err := registry.Databases(ctx)
		if err != nil {
			t.Fatal(err)
		}

		found := make(map[string]bool)
		for _, item := range items {
			found[item.Name] = true
		}
		if !found["pgo_mcp_test"] || !found["pgo_mcp_test_2"] {
			t.Fatalf("expected test databases, got %v", found)
		}
		if found["pgo_mcp_denied"] || found["template0"] || found["template1"] {
			t.Fatalf("ineligible database discovered: %v", found)
		}
	})

	t.Run("case sensitive allowlist", func(t *testing.T) {
		_, err := database.NewRegistry(
			ctx,
			uri,
			database.RegistryOptions{
				Discovery: true,
				Allowed:   []string{"PGO_MCP_TEST_2"},
				MaxPools:  2,
				MaxConns:  2,
			},
		)
		if err == nil {
			t.Fatal("expected case-sensitive allowlist error")
		}
	})

	t.Run("missing CONNECT privilege", func(t *testing.T) {
		_, err := database.NewRegistry(
			ctx,
			uri,
			database.RegistryOptions{
				Discovery: true,
				Allowed:   []string{"pgo_mcp_denied"},
				MaxPools:  2,
				MaxConns:  2,
			},
		)
		if err == nil {
			t.Fatal("expected CONNECT privilege error")
		}
	})
}

func TestSchemaAndObjects(t *testing.T) {
	ctx, session := testSession(t)

	schemas := callTool(
		t, ctx, session, "list_schemas",
		map[string]any{},
	)
	requireNonEmptyArray(t, schemas, "schemas")

	objects := callTool(
		t, ctx, session, "list_objects",
		map[string]any{
			"schema_name": "public",
			"object_type": "table",
		},
	)
	requireNonEmptyArray(t, objects, "objects")

	details := callTool(
		t, ctx, session, "get_object_details",
		map[string]any{
			"schema_name": "public",
			"object_name": "procurements",
			"object_type": "table",
		},
	)
	requireNonEmptyArray(t, details, "columns")
	requireNonEmptyArray(t, details, "constraints")
	requireNonEmptyArray(t, details, "indexes")
}

func TestExecuteSQL(t *testing.T) {
	ctx, session := testSession(t)

	result := callTool(
		t, ctx, session, "execute_sql",
		map[string]any{
			"sql": "SELECT current_database() AS db",
		},
	)

	rows, ok := result["rows"].([]any)
	if !ok || len(rows) != 1 {
		t.Fatalf("unexpected rows: %v", result["rows"])
	}

	first, ok := rows[0].([]any)
	if !ok || len(first) != 1 || first[0] != "pgo_mcp_test" {
		t.Fatalf("wrong database: %v", rows)
	}
}

func TestRowLimit(t *testing.T) {
	ctx, session := testSession(t)

	result := callTool(
		t, ctx, session, "execute_sql",
		map[string]any{
			"sql": "SELECT generate_series(1, 20)",
		},
	)

	if result["row_count"] != float64(5) {
		t.Fatalf(
			"expected QUERY_MAX_ROWS=5, running MCP server returned %v rows; "+
				"restart the server with QUERY_MAX_ROWS=5",
			result["row_count"],
		)
	}

	if result["truncated"] != true {
		t.Fatal("expected truncated=true")
	}
}

func TestReadOnly(t *testing.T) {
	ctx, session := testSession(t)

	requireToolError(
		t, ctx, session, "execute_sql",
		map[string]any{
			"sql": "CREATE TABLE public.should_not_exist (id int)",
		},
	)

	result := callTool(
		t, ctx, session, "execute_sql",
		map[string]any{
			"sql": "SELECT to_regclass('public.should_not_exist') AS obj",
		},
	)

	rows, ok := result["rows"].([]any)
	if !ok || len(rows) != 1 {
		t.Fatalf("unexpected rows: %v", result["rows"])
	}

	first, ok := rows[0].([]any)
	if !ok || len(first) != 1 || first[0] != nil {
		t.Fatalf("table unexpectedly exists: %v", rows)
	}

	requireToolError(
		t, ctx, session, "execute_sql",
		map[string]any{
			"sql": "SELECT 1; CREATE TABLE public.multi_statement_write (id int)",
		},
	)

	requireToolError(
		t, ctx, session, "execute_sql",
		map[string]any{
			"sql": "SELECT nextval('public.procurements_id_seq')",
		},
	)

	requireToolError(
		t, ctx, session, "explain_query",
		map[string]any{
			"sql":     "INSERT INTO public.procurements (status, amount) VALUES ('unsafe', 1)",
			"analyze": true,
		},
	)

	result = callTool(
		t, ctx, session, "execute_sql",
		map[string]any{
			"sql": "SELECT count(*) FROM public.procurements WHERE status = 'unsafe'",
		},
	)

	rows, ok = result["rows"].([]any)
	if !ok || len(rows) != 1 {
		t.Fatalf("unexpected rows: %v", result["rows"])
	}

	first, ok = rows[0].([]any)
	if !ok || len(first) != 1 || first[0] != "0" {
		t.Fatalf("write unexpectedly succeeded: %v", rows)
	}
}

func TestQueryLimits(t *testing.T) {
	ctx, session := testSession(t)

	result := callTool(
		t, ctx, session, "execute_sql",
		map[string]any{
			"sql": "SELECT repeat('x', 1048576)",
		},
	)

	if result["row_count"] != float64(0) || result["truncated"] != true {
		t.Fatalf("expected byte-limited result, got %v", result)
	}

	requireToolError(
		t, ctx, session, "execute_sql",
		map[string]any{
			"sql": "SELECT 1 /*" + strings.Repeat("x", 64*1024) + "*/",
		},
	)
}

func TestQueryCancellationAndCleanup(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URI"))
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("connect to test database: %v", err)
	}

	service := queryservice.New(
		pool,
		queryservice.Options{
			Timeout:  100 * time.Millisecond,
			MaxRows:  5,
			MaxBytes: 1024,
		},
	)

	started := time.Now()
	if _, err := service.ExecuteSQL(ctx, "SELECT pg_sleep(2)"); err == nil {
		t.Fatal("expected query timeout")
	}
	if time.Since(started) >= time.Second {
		t.Fatal("query cancellation exceeded expected deadline")
	}

	deadline := time.Now().Add(2 * time.Second)
	for pool.Stat().AcquiredConns() != 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if pool.Stat().AcquiredConns() != 0 {
		t.Fatalf(
			"expected no acquired connections, got %d",
			pool.Stat().AcquiredConns(),
		)
	}

	if _, err := service.ExecuteSQL(ctx, "SELECT 1"); err != nil {
		t.Fatalf("pool unusable after cancellation: %v", err)
	}
}

func TestExplainAndIndexAnalysis(t *testing.T) {
	ctx, session := testSession(t)

	sql := "SELECT * FROM public.procurements WHERE status = 'active'"

	explain := callTool(
		t, ctx, session, "explain_query",
		map[string]any{
			"sql":     sql,
			"analyze": false,
		},
	)
	requireNonEmptyArray(t, explain, "plan")

	indexes := callTool(
		t, ctx, session, "analyze_query_indexes",
		map[string]any{"sql": sql},
	)
	requireNonEmptyArray(t, indexes, "scans")
	requireNonEmptyArray(t, indexes, "tables")
}

func TestMonitoringTools(t *testing.T) {
	ctx, session := testSession(t)

	workload := callTool(
		t, ctx, session, "analyze_workload_indexes",
		map[string]any{"limit": 10},
	)
	if workload["database"] != "pgo_mcp_test" {
		t.Fatalf("wrong workload database: %v", workload)
	}

	health := callTool(
		t, ctx, session, "analyze_db_health",
		map[string]any{},
	)
	if health["database"] != "pgo_mcp_test" {
		t.Fatalf("wrong health database: %v", health)
	}

	top := callTool(
		t, ctx, session, "get_top_queries",
		map[string]any{
			"sort_by": "calls",
			"limit":   10,
		},
	)
	if top["database"] != "pgo_mcp_test" {
		t.Fatalf("wrong top queries database: %v", top)
	}
}

func TestInvalidInputs(t *testing.T) {
	ctx, session := testSession(t)

	tests := []struct {
		name string
		tool string
		args map[string]any
	}{
		{
			name: "invalid object type",
			tool: "list_objects",
			args: map[string]any{
				"schema_name": "public",
				"object_type": "invalid",
			},
		},
		{
			name: "invalid SQL",
			tool: "execute_sql",
			args: map[string]any{
				"sql": "SELECT * FROM missing_table",
			},
		},
		{
			name: "invalid sorting",
			tool: "get_top_queries",
			args: map[string]any{
				"sort_by": "invalid",
			},
		},
		{
			name: "invalid workload limit",
			tool: "analyze_workload_indexes",
			args: map[string]any{
				"limit": 100,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			requireToolError(
				t, ctx, session, tc.tool, tc.args,
			)
		})
	}
}

func TestHealthEndpoint(t *testing.T) {
	url := strings.TrimSuffix(endpoint(), "/mcp") + "/health"

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatal(fmt.Errorf(
			"health returned HTTP %d",
			resp.StatusCode,
		))
	}
}
