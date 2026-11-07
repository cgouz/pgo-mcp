package handler

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/cgouz/pgo-mcp/src/modules/query/dto"
	"github.com/cgouz/pgo-mcp/src/modules/query/service"
)

type Handler struct {
	service *service.Service
}

func New(s *service.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) ExecuteSQL(ctx context.Context, req *mcp.CallToolRequest, input dto.ExecuteSQLInput) (*mcp.CallToolResult, *dto.ExecuteSQLOutput, error) {

	result, err := h.service.ExecuteSQL(ctx, input.SQL)
	{
		if err != nil {
			return nil, nil, err
		}
	}

	return nil, result, nil
}

func (h *Handler) ExplainQuery(ctx context.Context, req *mcp.CallToolRequest, input dto.ExplainQueryInput) (*mcp.CallToolResult, *dto.ExplainQueryOutput, error) {

	result, err := h.service.ExplainQuery(ctx, input)
	{
		if err != nil {
			return nil, nil, err
		}
	}

	return nil, result, nil
}

func (h *Handler) AnalyzeQueryIndexes(ctx context.Context, req *mcp.CallToolRequest, input dto.AnalyzeQueryIndexesInput) (*mcp.CallToolResult, *dto.AnalyzeQueryIndexesOutput, error) {

	result, err := h.service.AnalyzeQueryIndexes(ctx, input)
	{
		if err != nil {
			return nil, nil, err
		}
	}

	return nil, result, nil
}

func (h *Handler) GetTopQueries(ctx context.Context, req *mcp.CallToolRequest, input dto.GetTopQueriesInput) (*mcp.CallToolResult, *dto.GetTopQueriesOutput, error) {

	result, err := h.service.GetTopQueries(ctx, input)
	{
		if err != nil {
			return nil, nil, err
		}
	}

	return nil, result, nil
}
