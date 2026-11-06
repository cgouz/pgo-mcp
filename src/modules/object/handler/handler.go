package handler

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/cgouz/pgo-mcp/src/modules/object/dto"
	"github.com/cgouz/pgo-mcp/src/modules/object/service"
)

type Handler struct {
	service *service.Service
}

func New(s *service.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) ListObjects(ctx context.Context, req *mcp.CallToolRequest, input dto.ListObjectsInput) (*mcp.CallToolResult, dto.ListObjectsOutput, error) {

	objectType := input.ObjectType

	if objectType == "" {
		objectType = "table"
	}

	objects, err := h.service.ListObjects(ctx, input.SchemaName, objectType)
	{
		if err != nil {
			return nil, dto.ListObjectsOutput{}, err
		}
	}

	return nil, dto.ListObjectsOutput{
		SchemaName: input.SchemaName,
		ObjectType: objectType,
		Objects:    objects,
	}, nil
}
