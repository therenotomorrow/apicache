package apiv1delete

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/therenotomorrow/apicache/internal/api"
	"github.com/therenotomorrow/apicache/internal/domain"
	"github.com/therenotomorrow/apicache/pkg/blender"
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
func Delete(cache domain.CacheDeleter) echo.HandlerFunc {
	params := blender.New[api.Params]()
	useCase := domain.NewDelUseCase(cache)

	return func(etx echo.Context) error {
		params, err := params.Path(etx)
		if err != nil {
			return api.UnprocessableEntityError(err)
		}

		err = useCase.Execute(etx.Request().Context(), params.Key)
		if err == nil {
			return etx.NoContent(http.StatusNoContent)
		}

		switch {
		case errors.Is(err, domain.ErrConnTimeout):
			return api.TooManyRequestsError(err)
		case errors.Is(err, domain.ErrContextTimeout):
			return api.TooManyRequestsError(err)
		}

		etx.Logger().Error(err)

		return api.InternalServerError(err)
	}
}
