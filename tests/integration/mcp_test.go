package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
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
		t.Fatalf("%s returned tool error: %+v", name, result.Content)
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
			"expected 5 rows, got %v",
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
