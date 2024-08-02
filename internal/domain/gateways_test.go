package domain_test

import (
	"testing"

	"github.com/therenotomorrow/apicache/internal/domain"
)

func TestUnitCacheGetter(t *testing.T) {
	t.Parallel()

	var _ domain.CacheGetter = getter{}
}

func TestUnitCacheSetter(t *testing.T) {
	t.Parallel()

	var _ domain.CacheSetter = setter{}
}

func TestUnitCacheDeleter(t *testing.T) {
	t.Parallel()

	var _ domain.CacheDeleter = deleter{}
}
