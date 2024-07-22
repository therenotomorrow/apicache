package apiv1post

import (
	"errors"
	"net/http"
	"time"

	"github.com/kxnes/go-interviews/apicache/internal/api"
	"github.com/kxnes/go-interviews/apicache/pkg/blender"
	"github.com/kxnes/go-interviews/apicache/pkg/cache"
	"github.com/labstack/echo/v4"
)

type (
	Params struct {
		Key string `param:"key" validate:"required"`
	}
	Payload struct {
		Val cache.ValType `json:"val" validate:"required"`
		TTL int           `json:"ttl" validate:"omitempty,min=0"`
	}
	Response struct {
		Key string        `json:"key"`
		Val cache.ValType `json:"val"`
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
func Post(cacher *cache.Cache) echo.HandlerFunc {
	params := blender.New[Params]()
	payload := blender.New[Payload]()

	return func(etx echo.Context) error {
		params, err := params.Path(etx)
		if err != nil {
			return api.UnprocessableEntityError(err)
		}

		payload, err := payload.JSON(etx)
		if err != nil {
			return api.UnprocessableEntityError(err)
		}

		err = cacher.Set(etx.Request().Context(), params.Key, payload.Val, time.Duration(payload.TTL)*time.Second)
		if err == nil {
			return etx.JSON(http.StatusCreated, &Response{Key: params.Key, Val: payload.Val})
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
