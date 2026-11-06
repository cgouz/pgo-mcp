package handler

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/cgouz/pgo-mcp/src/modules/schema/dto"
	"github.com/cgouz/pgo-mcp/src/modules/schema/service"
)

type Handler struct {
	service *service.Service
}

func New(s *service.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) ListSchemas(ctx context.Context, req *mcp.CallToolRequest, input dto.ListSchemasInput) (*mcp.CallToolResult, dto.ListSchemasOutput, error) {

	schemas, err := h.service.ListSchemas(ctx)
	{
		if err != nil {
			return nil, dto.ListSchemasOutput{}, err
		}
	}

	return nil, dto.ListSchemasOutput{
		Schemas: schemas,
	}, nil
}
