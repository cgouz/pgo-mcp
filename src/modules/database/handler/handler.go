package handler

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/cgouz/pgo-mcp/src/modules/database/dto"
	"github.com/cgouz/pgo-mcp/src/modules/database/service"
)

type Handler struct {
	service *service.Service
}

func New(s *service.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) ListDatabases(ctx context.Context, req *mcp.CallToolRequest, input dto.ListDatabasesInput) (*mcp.CallToolResult, dto.ListDatabasesOutput, error) {

	result, err := h.service.ListDatabases(ctx)
	if err != nil {
		return nil, dto.ListDatabasesOutput{}, err
	}

	return nil, result, nil
}
