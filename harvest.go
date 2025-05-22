package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/isaquerr25/go-templ-htmx/views/pages/harvest"
	"github.com/labstack/echo/v4"
)

func ListHarvest(c echo.Context) error {
	var harvests []Harvest
	if err := db.Find(&harvests).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Erro ao buscar colheitas")
	}

	var items []harvest.HarvestProps
	for _, h := range harvests {
		items = append(items, harvest.HarvestProps{
			ID:          h.ID,
			PlantingID:  h.PlantingID,
			HarvestedAt: h.HarvestedAt.Format("2006-01-02"),
			Quantity:    h.Quantity,
			Unit:        h.Unit,
			SaleValue:   h.SaleValue,
		})
	}

	return harvest.List(harvest.HarvestListProps{Items: items}).
		Render(c.Request().Context(), c.Response())
}

func ShowHarvest(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var h Harvest
	if err := db.First(&h, id).Error; err != nil {
		return c.String(http.StatusNotFound, "Colheita não encontrada")
	}

	p := harvest.HarvestProps{
		ID:          h.ID,
		PlantingID:  h.PlantingID,
		HarvestedAt: h.HarvestedAt.Format("2006-01-02"),
		Quantity:    h.Quantity,
		Unit:        h.Unit,
		SaleValue:   h.SaleValue,
	}

	return harvest.Index(p).Render(c.Request().Context(), c.Response())
}

func CreateHarvest(c echo.Context) error {
	var input harvest.HarvestProps
	if err := c.Bind(&input); err != nil {
		return c.String(http.StatusBadRequest, "Erro ao processar dados")
	}

	date, _ := time.Parse("2006-01-02", c.FormValue("harvestedAt"))

	h := Harvest{
		PlantingID:  input.PlantingID,
		HarvestedAt: date,
		Quantity:    input.Quantity,
		Unit:        input.Unit,
		SaleValue:   input.SaleValue,
	}

	if err := db.Create(&h).Error; err != nil {
		input.Error = map[string]string{"global": "Erro ao salvar"}
		return harvest.Index(input).Render(c.Request().Context(), c.Response())
	}

	return ListHarvest(c)
}

func UpdateHarvest(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var h Harvest
	if err := db.First(&h, id).Error; err != nil {
		return c.String(http.StatusNotFound, "Colheita não encontrada")
	}

	var input harvest.HarvestProps
	if err := c.Bind(&input); err != nil {
		return c.String(http.StatusBadRequest, "Erro no bind")
	}

	date, _ := time.Parse("2006-01-02", c.FormValue("harvestedAt"))

	h.PlantingID = input.PlantingID
	h.HarvestedAt = date
	h.Quantity = input.Quantity
	h.Unit = input.Unit
	h.SaleValue = input.SaleValue

	if err := db.Save(&h).Error; err != nil {
		input.Error = map[string]string{"global": "Erro ao atualizar"}
		return harvest.Index(input).Render(c.Request().Context(), c.Response())
	}

	return ListHarvest(c)
}
