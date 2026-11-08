package handler

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/cgouz/pgo-mcp/src/infrastructure/database"
	"github.com/cgouz/pgo-mcp/src/modules/index/dto"
	"github.com/cgouz/pgo-mcp/src/modules/index/service"
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

func (h *Handler) AnalyzeWorkloadIndexes(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input dto.AnalyzeWorkloadInput,
) (*mcp.CallToolResult, dto.AnalyzeWorkloadOutput, error) {
	pool, err := h.registry.Pool(ctx, input.Database)
	if err != nil {
		return nil, dto.AnalyzeWorkloadOutput{}, err
	}

	result, err := service.New(pool, h.options).AnalyzeWorkload(ctx, input)
	if err != nil {
		return nil, dto.AnalyzeWorkloadOutput{}, err
	}

	return nil, *result, nil
}
