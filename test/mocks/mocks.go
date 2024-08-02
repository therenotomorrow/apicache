package mocks

import "context"

type DriverMock struct {
	CloseMock func() error
	GetMock   func(ctx context.Context, key string) (string, error)
	SetMock   func(ctx context.Context, key string, val string) error
	DelMock   func(ctx context.Context, key string) error
}

func NewDriverMock() *DriverMock {
	return &DriverMock{
		CloseMock: func() error { return nil },
		GetMock:   nil,
		SetMock:   func(_ context.Context, _ string, _ string) error { return nil },
		DelMock:   func(_ context.Context, _ string) error { return nil },
	}
}

func (d *DriverMock) Get(ctx context.Context, key string) (string, error) { return d.GetMock(ctx, key) }
func (d *DriverMock) Set(ctx context.Context, key string, val string) error {
	return d.SetMock(ctx, key, val)
}
func (d *DriverMock) Del(ctx context.Context, key string) error { return d.DelMock(ctx, key) }
func (d *DriverMock) Close() error                              { return d.CloseMock() }
