package swagger

import (
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func Connect(router *echo.Echo) {
	router.GET("/api/docs/*", echoSwagger.WrapHandler)
}
