package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/isaquerr25/go-templ-htmx/views/pages/fertilization"
	"github.com/labstack/echo/v4"
)

func ListFertilization(c echo.Context) error {
	var ferts []Fertilization
	if err := db.Find(&ferts).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Erro ao buscar fertilizações")
	}

	var items []fertilization.FertilizationProps
	for _, f := range ferts {
		items = append(items, fertilization.FertilizationProps{
			ID:              f.ID,
			PlantingID:      f.PlantingID,
			ProductID:       f.ProductID,
			ApplicationType: f.ApplicationType,
			AppliedAt:       f.AppliedAt,
			QuantityUsed:    f.QuantityUsed,
			Unit:            f.Unit,
		})
	}

	return fertilization.List(fertilization.FertilizationListProps{Items: items}).
		Render(c.Request().Context(), c.Response())
}

func ShowFertilization(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	var f Fertilization
	if err := db.First(&f, id).Error; err != nil {
		return c.String(http.StatusNotFound, "Fertilização não encontrada")
	}

	p := fertilization.FertilizationProps{
		ID:              f.ID,
		PlantingID:      f.PlantingID,
		ProductID:       f.ProductID,
		ApplicationType: f.ApplicationType,
		AppliedAt:       f.AppliedAt,
		QuantityUsed:    f.QuantityUsed,
		Unit:            f.Unit,
	}

	return fertilization.Index(p).Render(c.Request().Context(), c.Response())
}

func CreateFertilization(c echo.Context) error {
	var input fertilization.FertilizationProps
	if err := c.Bind(&input); err != nil {
		return c.String(http.StatusBadRequest, "Erro no bind")
	}

	appliedAt, _ := time.Parse("2006-01-02", c.FormValue("appliedAt"))

	f := Fertilization{
		PlantingID:      input.PlantingID,
		ProductID:       input.ProductID,
		ApplicationType: input.ApplicationType,
		AppliedAt:       appliedAt,
		QuantityUsed:    input.QuantityUsed,
		Unit:            input.Unit,
	}

	if err := db.Create(&f).Error; err != nil {
		input.Error = map[string]string{"global": "Erro ao salvar"}
		return fertilization.Index(input).Render(c.Request().Context(), c.Response())
	}

	return ListFertilization(c)
}

func UpdateFertilization(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	var f Fertilization
	if err := db.First(&f, id).Error; err != nil {
		return c.String(http.StatusNotFound, "Fertilização não encontrada")
	}

	var input fertilization.FertilizationProps
	if err := c.Bind(&input); err != nil {
		return c.String(http.StatusBadRequest, "Erro no bind")
	}

	appliedAt, _ := time.Parse("2006-01-02", c.FormValue("appliedAt"))

	f.PlantingID = input.PlantingID
	f.ProductID = input.ProductID
	f.ApplicationType = input.ApplicationType
	f.AppliedAt = appliedAt
	f.QuantityUsed = input.QuantityUsed
	f.Unit = input.Unit

	if err := db.Save(&f).Error; err != nil {
		input.Error = map[string]string{"global": "Erro ao atualizar"}
		return fertilization.Index(input).Render(c.Request().Context(), c.Response())
	}

	return ListFertilization(c)
}
