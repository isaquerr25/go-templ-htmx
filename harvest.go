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
			ID:         h.ID,
			PlantingID: h.PlantingID,
			HarvestedAt: harvest.Date{
				Time: h.HarvestedAt,
			},
			Quantity: h.Quantity,
			Unit:     h.Unit,
			Error:    map[string]string{},
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
		ID:         h.ID,
		PlantingID: h.PlantingID,
		HarvestedAt: harvest.Date{
			Time: h.HarvestedAt,
		},
		Quantity: h.Quantity,
		Unit:     h.Unit,
	}

	return harvest.Index(p).Render(c.Request().Context(), c.Response())
}

func CreateHarvest() echo.HandlerFunc {
	return func(c echo.Context) error {
		planId := c.Param("planId")

		// Extrair os campos
		quantityStr := c.FormValue("quantity")
		unit := c.FormValue("unit")
		harvestedAtStr := c.FormValue("appliedAt")

		// Parse de tipos
		plantingId, err := strconv.Atoi(planId)
		if err != nil {
			return c.String(http.StatusBadRequest, "PlantingID inválido")
		}

		quantity, err := strconv.ParseFloat(quantityStr, 64)
		if err != nil {
			return c.String(http.StatusBadRequest, "Quantidade inválida")
		}

		harvestedAt, err := time.Parse("2006-01-02", harvestedAtStr)
		if err != nil {
			return c.String(http.StatusBadRequest, "Data (HarvestedAt) inválida")
		}

		// Criar o registro de colheita
		h := Harvest{
			PlantingID:  uint(plantingId),
			HarvestedAt: harvestedAt,
			Quantity:    quantity,
			Unit:        unit,
		}

		if err := db.Create(&h).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao salvar colheita")
		}

		c.Response().Header().Set("HX-Redirect", "../")
		return c.String(http.StatusOK, "")
	}
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

	if err := db.Save(&h).Error; err != nil {
		input.Error = map[string]string{"global": "Erro ao atualizar"}
		return harvest.Index(input).Render(c.Request().Context(), c.Response())
	}

	return ListHarvest(c)
}
