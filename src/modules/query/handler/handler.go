package handler

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/cgouz/pgo-mcp/src/infrastructure/database"
	"github.com/cgouz/pgo-mcp/src/modules/query/dto"
	"github.com/cgouz/pgo-mcp/src/modules/query/service"
)

type Handler struct {
	registry *database.Registry
	options  service.Options
}

func New(registry *database.Registry, options service.Options) *Handler {
	return &Handler{
		registry: registry,
		options:  options,
	}
}

func (h *Handler) ExecuteSQL(ctx context.Context, req *mcp.CallToolRequest, input dto.ExecuteSQLInput) (*mcp.CallToolResult, *dto.ExecuteSQLOutput, error) {

	pool, err := h.registry.Pool(ctx, input.Database)
	{
		if err != nil {
			return nil, nil, err
		}
	}

	result, err := service.New(pool, h.options).ExecuteSQL(ctx, input.SQL)
	{
		if err != nil {
			return nil, nil, err
		}
	}

	return nil, result, nil
}

func (h *Handler) ExplainQuery(ctx context.Context, req *mcp.CallToolRequest, input dto.ExplainQueryInput) (*mcp.CallToolResult, *dto.ExplainQueryOutput, error) {

	pool, err := h.registry.Pool(ctx, input.Database)
	{
		if err != nil {
			return nil, nil, err
		}
	}

	result, err := service.New(pool, h.options).ExplainQuery(ctx, input)
	{
		if err != nil {
			return nil, nil, err
		}
	}

	return nil, result, nil
}

func (h *Handler) AnalyzeQueryIndexes(ctx context.Context, req *mcp.CallToolRequest, input dto.AnalyzeQueryIndexesInput) (*mcp.CallToolResult, *dto.AnalyzeQueryIndexesOutput, error) {

	pool, err := h.registry.Pool(ctx, input.Database)
	{
		if err != nil {
			return nil, nil, err
		}
	}

	result, err := service.New(pool, h.options).AnalyzeQueryIndexes(ctx, input)
	{
		if err != nil {
			return nil, nil, err
		}
	}

	return nil, result, nil
}

func (h *Handler) GetTopQueries(ctx context.Context, req *mcp.CallToolRequest, input dto.GetTopQueriesInput) (*mcp.CallToolResult, *dto.GetTopQueriesOutput, error) {

	pool, err := h.registry.Pool(ctx, input.Database)
	{
		if err != nil {
			return nil, nil, err
		}
	}

	result, err := service.New(pool, h.options).GetTopQueries(ctx, input)
	{
		if err != nil {
			return nil, nil, err
		}
	}

	return nil, result, nil
}
