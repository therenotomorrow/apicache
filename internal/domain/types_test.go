package domain_test

import (
	"testing"

	"github.com/kxnes/go-interviews/apicache/internal/domain"
)

func TestUnitValType(t *testing.T) {
	t.Parallel()

	var _ domain.ValType = map[string]any{"hello": "world", "age": 42}
}
