package apiv1get

import (
	"errors"
	"net/http"

	"github.com/kxnes/go-interviews/apicache/internal/api"
	"github.com/kxnes/go-interviews/apicache/internal/domain"
	"github.com/kxnes/go-interviews/apicache/pkg/blender"
	"github.com/labstack/echo/v4"
)

type Response struct {
	Key string         `json:"key"`
	Val domain.ValType `json:"val"`
}

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
func Get(cache domain.CacheGetter) echo.HandlerFunc {
	params := blender.New[api.Params]()
	useCase := domain.NewGetUseCase(cache)

	return func(etx echo.Context) error {
		params, err := params.Path(etx)
		if err != nil {
			return api.UnprocessableEntityError(err)
		}

		val, err := useCase.Execute(etx.Request().Context(), params.Key)
		if err == nil {
			return etx.JSON(http.StatusOK, &Response{Key: params.Key, Val: val})
		}

		switch {
		case errors.Is(err, domain.ErrKeyExpired):
			return api.BadRequestError(err)
		case errors.Is(err, domain.ErrKeyNotExist):
			return api.NotFoundError(err)
		case errors.Is(err, domain.ErrConnTimeout):
			return api.TooManyRequestsError(err)
		case errors.Is(err, domain.ErrContextTimeout):
			return api.TooManyRequestsError(err)
		}

		etx.Logger().Error(err)

		return api.InternalServerError(err)
	}
}
