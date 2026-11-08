package service

import (
	"context"

	"github.com/cgouz/pgo-mcp/src/infrastructure/database"
	"github.com/cgouz/pgo-mcp/src/modules/database/dto"
)

type Service struct {
	registry *database.Registry
}

func New(registry *database.Registry) *Service {
	return &Service{registry: registry}
}

func (s *Service) ListDatabases(ctx context.Context) (dto.ListDatabasesOutput, error) {

	databases, err := s.registry.Databases(ctx)
	if err != nil {
		return dto.ListDatabasesOutput{}, err
	}

	result := dto.ListDatabasesOutput{
		Default:   s.registry.DefaultName(),
		Databases: make([]dto.Database, 0, len(databases)),
	}

	for _, item := range databases {
		result.Databases = append(result.Databases, dto.Database{
			Name:    item.Name,
			Default: item.Default,
		})
	}

	return result, nil
}
