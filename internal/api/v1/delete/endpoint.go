package apiv1delete

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
	CacheDeleter interface {
		Del(ctx context.Context, key string) error
	}
	Params struct {
		Key string `param:"key" validate:"required"`
	}
)

// Delete ----
// @Summary    "Delete key/value pair"
// @Tags       cache
// @Param      key path string true "Key"
// @Success    204
// @Failure    422 {object} api.UnprocessableEntity
// @Failure    429 {object} api.TooManyRequests
// @Failure    500 {object} api.InternalServer
// @Router     /api/v1/{key}/ [delete].
func Delete(cacher CacheDeleter) echo.HandlerFunc {
	params := blender.New[Params]()

	return func(etx echo.Context) error {
		params, err := params.Path(etx)
		if err != nil {
			return api.UnprocessableEntityError(err)
		}

		err = cacher.Del(etx.Request().Context(), params.Key)
		if err == nil {
			return etx.NoContent(http.StatusNoContent)
		}

		switch {
		case errors.Is(err, cache.ErrConnTimeout):
			return api.TooManyRequestsError(err)
		case errors.Is(err, cache.ErrContextTimeout):
			return api.TooManyRequestsError(err)
		}

		etx.Logger().Error(err)

		return api.InternalServerError(err)
	}
}
