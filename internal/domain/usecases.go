package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const defaultTTL = 0

type (
	GetUseCase struct {
		cache CacheGetter
	}
	SetUseCase struct {
		cache CacheSetter
	}
	DelUseCase struct {
		cache CacheDeleter
	}
)

func NewGetUseCase(cache CacheGetter) *GetUseCase {
	return &GetUseCase{cache: cache}
}

func (use *GetUseCase) Execute(ctx context.Context, key string) (ValType, error) {
	if key == "" {
		return nil, ErrEmptyKey
	}

	raw, err := use.cache.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	var val ValType

	err = json.Unmarshal(raw, &val)
	if err != nil {
		return nil, ErrDataCorrupted
	}

	return val, nil
}

func NewSetUseCase(cache CacheSetter) *SetUseCase {
	return &SetUseCase{cache: cache}
}

func (use *SetUseCase) Execute(ctx context.Context, key string, val ValType, ttl int) error {
	if key == "" {
		return ErrEmptyKey
	}

	if val == nil {
		return ErrEmptyVal
	}

	raw, err := json.Marshal(val)
	if err != nil {
		return ErrDataCorrupted
	}

	var deadline time.Time

	if ttl > defaultTTL {
		deadline = time.Now().UTC().Add(time.Duration(ttl) * time.Second)
	}

	err = use.cache.Set(ctx, key, raw, deadline)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func NewDelUseCase(cache CacheDeleter) *DelUseCase {
	return &DelUseCase{cache: cache}
}

func (use *DelUseCase) Execute(ctx context.Context, key string) error {
	if key == "" {
		return ErrEmptyKey
	}

	err := use.cache.Del(ctx, key)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
