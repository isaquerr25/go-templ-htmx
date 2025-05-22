package main

import (
	"fmt"

	"github.com/isaquerr25/go-templ-htmx/views/pages/produto"
	"github.com/labstack/echo/v4"
)

type Server struct{}

func (s Server) UpdateProduct(c echo.Context) error {
	p := &Product{}
	r := db.First(p, c.Param("ID"))
	if r.Error != nil {
		fmt.Println(r.Error)
		return r.Error
	}

	k, hasError, err := validateProduct(c, p)
	if err != nil {
		return err
	}

	if !hasError {
		r := db.Save(&p)
		if r.Error != nil {
			fmt.Println(r.Error)

			return r.Error
		}
		c.Response().Header().Set("HX-Redirect", "/listProduct")
		c.Response().WriteHeader(200)
		return c.String(200, "")
	}

	return Render(c, 200, produto.Index(k))
}
