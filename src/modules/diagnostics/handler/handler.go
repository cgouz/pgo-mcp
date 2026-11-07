package handler

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/cgouz/pgo-mcp/src/modules/diagnostics/dto"
	"github.com/cgouz/pgo-mcp/src/modules/diagnostics/service"
)

type Handler struct {
	service *service.Service
}

func New(s *service.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) AnalyzeDBHealth(ctx context.Context, req *mcp.CallToolRequest, input dto.AnalyzeDBHealthInput) (*mcp.CallToolResult, *dto.AnalyzeDBHealthOutput, error) {

	result, err := h.service.AnalyzeDBHealth(ctx)
	{
		if err != nil {
			return nil, nil, err
		}
	}

	return nil, result, nil
}
