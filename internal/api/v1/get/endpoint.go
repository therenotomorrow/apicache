package apiv1get

import (
	"context"
	"errors"
	"net/http"

	"github.com/kxnes/go-interviews/apicache/internal/api"
	"github.com/kxnes/go-interviews/apicache/pkg/blender"
	"github.com/kxnes/go-interviews/apicache/pkg/cache"
	"github.com/labstack/echo/v4"
)

type (
	CacheGetter interface {
		Get(ctx context.Context, key string) (cache.ValType, error)
	}
	Params struct {
		Key string `param:"key" validate:"required"`
	}
	Response struct {
		Key string        `json:"key"`
		Val cache.ValType `json:"val"`
	}
)

// Get ----
// @Summary    "Retrieve key/value pair"
// @Tags       cache
// @Param      key path string true "Key"
// @Produce    json
// @Success    200 {object} Response
// @Failure    400 {object} api.BadRequest
// @Failure    404 {object} api.NotFound
// @Failure    422 {object} api.UnprocessableEntity
// @Failure    429 {object} api.TooManyRequests
// @Failure    500 {object} api.InternalServer
// @Router     /api/v1/{key}/ [get].
func Get(cacher CacheGetter) echo.HandlerFunc {
	params := blender.New[Params]()

	return func(etx echo.Context) error {
		params, err := params.Path(etx)
		if err != nil {
			return api.UnprocessableEntityError(err)
		}

		val, err := cacher.Get(etx.Request().Context(), params.Key)
		if err == nil {
			return etx.JSON(http.StatusOK, &Response{Key: params.Key, Val: val})
		}

		switch {
		case errors.Is(err, cache.ErrKeyExpired):
			return api.BadRequestError(err)
		case errors.Is(err, cache.ErrKeyNotExist):
			return api.NotFoundError(err)
		case errors.Is(err, cache.ErrConnTimeout):
			return api.TooManyRequestsError(err)
		case errors.Is(err, cache.ErrContextTimeout):
			return api.TooManyRequestsError(err)
		}

		etx.Logger().Error(err)

		return api.InternalServerError(err)
	}
}
