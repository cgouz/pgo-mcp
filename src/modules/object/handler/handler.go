package handler

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/cgouz/pgo-mcp/src/infrastructure/database"
	"github.com/cgouz/pgo-mcp/src/modules/object/dto"
	"github.com/cgouz/pgo-mcp/src/modules/object/service"
)

type Handler struct {
	registry *database.Registry
}

func New(registry *database.Registry) *Handler {
	return &Handler{registry: registry}
}

func (h *Handler) ListObjects(ctx context.Context, req *mcp.CallToolRequest, input dto.ListObjectsInput) (*mcp.CallToolResult, dto.ListObjectsOutput, error) {

	objectType := input.ObjectType

	if objectType == "" {
		objectType = "table"
	}

	pool, err := h.registry.Pool(ctx, input.Database)
	{
		if err != nil {
			return nil, dto.ListObjectsOutput{}, err
		}
	}

	objects, err := service.New(pool).ListObjects(ctx, input.SchemaName, objectType)
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

func (h *Handler) GetObjectDetails(ctx context.Context, req *mcp.CallToolRequest, input dto.GetObjectDetailsInput) (*mcp.CallToolResult, dto.GetObjectDetailsOutput, error) {

	pool, err := h.registry.Pool(ctx, input.Database)
	{
		if err != nil {
			return nil, dto.GetObjectDetailsOutput{}, err
		}
	}

	details, err := service.New(pool).GetObjectDetails(ctx, input)
	{
		if err != nil {
			return nil, dto.GetObjectDetailsOutput{}, err
		}
	}

	return nil, details, nil
}
