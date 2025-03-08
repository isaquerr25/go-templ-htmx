package main

import (
	"embed"

	"github.com/a-h/templ"
	"github.com/isaquerr25/go-templ-htmx/views/pages/home"
	"github.com/labstack/echo/v4"
)

//go:embed static/*
var assets embed.FS

func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	if err := t.Render(ctx.Request().Context(), buf); err != nil {
		return err
	}

	return ctx.HTML(statusCode, buf.String())
}

func main() {
	e := echo.New()
	e.Static("/static", "static")

	e.GET("/",
		func(c echo.Context) error {
			return Render(c, 200, home.Hello("asds"))
		},
	)
	e.Logger.Fatal(e.Start(":1323"))
}
