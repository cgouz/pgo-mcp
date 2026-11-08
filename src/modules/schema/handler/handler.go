package handler

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/cgouz/pgo-mcp/src/infrastructure/database"
	"github.com/cgouz/pgo-mcp/src/modules/schema/dto"
	"github.com/cgouz/pgo-mcp/src/modules/schema/service"
)

type Handler struct {
	registry *database.Registry
}

func New(registry *database.Registry) *Handler {
	return &Handler{registry: registry}
}

func (h *Handler) ListSchemas(ctx context.Context, req *mcp.CallToolRequest, input dto.ListSchemasInput) (*mcp.CallToolResult, dto.ListSchemasOutput, error) {

	pool, err := h.registry.Pool(ctx, input.Database)
	{
		if err != nil {
			return nil, dto.ListSchemasOutput{}, err
		}
	}

	schemas, err := service.New(pool).ListSchemas(ctx)
	{
		if err != nil {
			return nil, dto.ListSchemasOutput{}, err
		}
	}

	return nil, dto.ListSchemasOutput{
		Schemas: schemas,
	}, nil
}
