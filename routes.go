package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// Message type
type Message struct {
	Content string `json:"content"`
}

func newRouter() *echo.Echo {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", UserGET)
	e.POST("/", UserPOST)
	e.PUT("/", UserPUT)
	e.DELETE("/", UserDELETE)

	return e
}
