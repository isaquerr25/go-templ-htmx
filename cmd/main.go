package main

import (
	"context"
	"github.com/isaquerr25/go-templ-htmx/handler"
	"github.com/labstack/echo/v4"
)

func main() {
	app := echo.New()

	userHandler := handler.UserHandler{}
	app.Use(withUser)
	app.GET("/user", userHandler.HandleUserShow)
	app.Start(":3000")
}

func withUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := context.WithValue(c.Request().Context(), "user", "dasdasd@dasd.das")
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}
