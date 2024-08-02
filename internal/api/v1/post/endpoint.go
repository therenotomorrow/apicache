package apiv1post

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/therenotomorrow/apicache/internal/api"
	"github.com/therenotomorrow/apicache/internal/domain"
	"github.com/therenotomorrow/apicache/pkg/blender"
)

type (
	Payload struct {
		Val domain.ValType `json:"val" validate:"required"`
		TTL int            `json:"ttl" validate:"omitempty,min=0"`
	}
	Response struct {
		Key string         `json:"key"`
		Val domain.ValType `json:"val"`
	}
)

// Post ----
// @Summary    "Insert key/value pair"
// @Tags       cache
// @Param      key path string true "Key"
// @Accept     json
// @Param      payload body Payload true "Payload"
// @Produce    json
// @Success    201 {object} Response
// @Failure    422 {object} api.UnprocessableEntity
// @Failure    429 {object} api.TooManyRequests
// @Failure    500 {object} api.InternalServer
// @Router     /api/v1/{key}/ [post].
func Post(cache domain.CacheSetter) echo.HandlerFunc {
	params := blender.New[api.Params]()
	payload := blender.New[Payload]()
	useCase := domain.NewSetUseCase(cache)

	return func(etx echo.Context) error {
		params, err := params.Path(etx)
		if err != nil {
			return api.UnprocessableEntityError(err)
		}

		payload, err := payload.JSON(etx)
		if err != nil {
			return api.UnprocessableEntityError(err)
		}

		err = useCase.Execute(etx.Request().Context(), params.Key, payload.Val, payload.TTL)
		if err == nil {
			return etx.JSON(http.StatusCreated, &Response{Key: params.Key, Val: payload.Val})
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
