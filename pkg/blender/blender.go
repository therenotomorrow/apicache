package blender

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type Blender[T any] struct {
	binder   *echo.DefaultBinder
	validate *validator.Validate
}

func New[T any]() *Blender[T] {
	return &Blender[T]{binder: new(echo.DefaultBinder), validate: validator.New()}
}

func (b *Blender[T]) validateStruct(data *T) (*T, error) {
	err := b.validate.Struct(data)
	if err != nil {
		return nil, fmt.Errorf("validate error: %w", err)
	}

	return data, nil
}

func (b *Blender[T]) JSON(etx echo.Context) (*T, error) {
	payload := new(T)
	// echo doesn't close the body of the request
	defer func() { _ = etx.Request().Body.Close() }()

	err := b.binder.BindBody(etx, payload)
	if err != nil {
		return nil, fmt.Errorf("json error: %w", err)
	}

	return b.validateStruct(payload)
}

func (b *Blender[T]) Path(etx echo.Context) (*T, error) {
	params := new(T)

	err := b.binder.BindPathParams(etx, params)
	if err != nil {
		return nil, fmt.Errorf("path error: %w", err)
	}

	return b.validateStruct(params)
}
