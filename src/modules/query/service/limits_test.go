package service

import (
	"context"
	"strings"
	"testing"

	"github.com/cgouz/pgo-mcp/src/modules/query/dto"
)

func TestSQLSizeLimit(t *testing.T) {
	service := New(nil, Options{})
	sql := strings.Repeat("x", maxSQLBytes+1)

	if _, err := service.ExecuteSQL(context.Background(), sql); err == nil {
		t.Fatal("ExecuteSQL accepted oversized SQL")
	}

	if _, err := service.ExplainQuery(
		context.Background(),
		dto.ExplainQueryInput{SQL: sql},
	); err == nil {
		t.Fatal("ExplainQuery accepted oversized SQL")
	}
}
