package handler

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/cgouz/pgo-mcp/src/modules/index/dto"
	"github.com/cgouz/pgo-mcp/src/modules/index/service"
)

type Handler struct {
	service *service.Service
}

func New(s *service.Service) *Handler {
	return &Handler{
		service: s,
	}
}

func (h *Handler) AnalyzeWorkloadIndexes(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input dto.AnalyzeWorkloadInput,
) (*mcp.CallToolResult, dto.AnalyzeWorkloadOutput, error) {
	result, err := h.service.AnalyzeWorkload(ctx, input)
	if err != nil {
		return nil, dto.AnalyzeWorkloadOutput{}, err
	}

	return nil, *result, nil
}
