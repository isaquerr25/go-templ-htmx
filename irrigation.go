package main

import (
	"net/http"
	"strconv"

	"github.com/isaquerr25/go-templ-htmx/views/pages/irrigation"
	"github.com/labstack/echo/v4"
)

var (
	irrigations []irrigation.IrrigationSectorProps
	nextID      uint = 1
)

func IrrigationIndex(c echo.Context) error {
	return irrigation.Index(irrigation.IrrigationSectorProps{}).
		Render(c.Request().Context(), c.Response().Writer)
}

func IrrigationCreate(c echo.Context) error {
	var form irrigation.IrrigationSectorProps
	if err := c.Bind(&form); err != nil {
		return c.String(http.StatusBadRequest, "Erro ao bindar dados")
	}
	form.ID = nextID
	nextID++
	irrigations = append(irrigations, form)
	return IrrigationList(c)
}

func IrrigationList(c echo.Context) error {
	list := irrigation.IrrigationSectorListProps{Items: irrigations}
	return irrigation.List(list).Render(c.Request().Context(), c.Response().Writer)
}

func IrrigationEdit(c echo.Context) error {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	for _, item := range irrigations {
		if item.ID == uint(id) {
			return irrigation.Index(item).Render(c.Request().Context(), c.Response().Writer)
		}
	}
	return c.String(http.StatusNotFound, "Irrigação não encontrada")
}
