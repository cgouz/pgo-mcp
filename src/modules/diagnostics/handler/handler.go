package handler

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/cgouz/pgo-mcp/src/infrastructure/database"
	"github.com/cgouz/pgo-mcp/src/modules/diagnostics/dto"
	"github.com/cgouz/pgo-mcp/src/modules/diagnostics/service"
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

func (h *Handler) AnalyzeDBHealth(ctx context.Context, req *mcp.CallToolRequest, input dto.AnalyzeDBHealthInput) (*mcp.CallToolResult, *dto.AnalyzeDBHealthOutput, error) {

	pool, err := h.registry.Pool(ctx, input.Database)
	{
		if err != nil {
			return nil, nil, err
		}
	}

	result, err := service.New(pool, h.options).AnalyzeDBHealth(ctx)
	{
		if err != nil {
			return nil, nil, err
		}
	}

	return nil, result, nil
}
