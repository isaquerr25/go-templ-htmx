package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/isaquerr25/go-templ-htmx/views/pages/harvest"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func DeleteHarvest(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "ID inválido")
	}

	if err := db.Delete(&Harvest{}, id).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Erro ao deletar colheita")
	}

	// Se for uma requisição HTMX, pode responder com um redirect ou fragmento
	if c.Request().Header.Get("HX-Request") == "true" {
		c.Response().Header().Set("HX-Redirect", "")
		return c.String(http.StatusOK, "")
	}

	return c.Redirect(http.StatusSeeOther, "")
}

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

func CreateHarvest(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		fmt.Println("👉 Iniciando CreateHarvest")

		planId := c.Param("planId")
		fmt.Println("📌 planId recebido:", planId)

		// Parse IDs e campos
		plantingId, err := strconv.Atoi(planId)
		if err != nil {
			fmt.Println("❌ Erro ao converter planId:", err)
			return c.String(http.StatusBadRequest, "PlantingID inválido")
		}

		quantityStr := c.FormValue("quantity")
		unit := c.FormValue("unit")
		harvestedAtStr := c.FormValue("appliedAt")

		fmt.Println("📥 quantityStr:", quantityStr)
		fmt.Println("📥 unit:", unit)
		fmt.Println("📥 harvestedAtStr:", harvestedAtStr)

		quantity, err := strconv.ParseFloat(quantityStr, 64)
		if err != nil {
			fmt.Println("❌ Erro ao converter quantidade:", err)
			return c.String(http.StatusBadRequest, "Quantidade inválida")
		}

		harvestedAt, err := time.Parse("2006-01-02", harvestedAtStr)
		if err != nil {
			fmt.Println("❌ Erro ao converter data:", err)
			return c.String(http.StatusBadRequest, "Data (HarvestedAt) inválida")
		}

		// Criar colheita
		h := Harvest{
			PlantingID:  uint(plantingId),
			HarvestedAt: harvestedAt,
			Quantity:    quantity,
			Unit:        unit,
		}

		fmt.Println("✅ Colheita a ser salva:", h)

		if err := db.Create(&h).Error; err != nil {
			fmt.Println("❌ Erro ao salvar colheita:", err)
			return c.String(http.StatusInternalServerError, "Erro ao salvar colheita")
		}

		// Buscar o Planting para pegar o TypeProductID
		var planting Planting
		if err := db.First(&planting, plantingId).Error; err != nil {
			fmt.Println("❌ Erro ao buscar plantio:", err)
			return c.String(http.StatusInternalServerError, "Erro ao buscar plantio")
		}
		fmt.Println("🌱 Plantio encontrado:", planting)

		// Atualizar valor no TypeProduct
		var tp TypeProduct
		if err := db.First(&tp, planting.TypeProductID).Error; err != nil {
			fmt.Println("❌ Erro ao buscar tipo de produto:", err)
			return c.String(http.StatusInternalServerError, "Erro ao buscar tipo de produto")
		}
		fmt.Println("📦 TypeProduct antes da atualização:", tp)

		tp.Quantity += quantity

		if err := db.Save(&tp).Error; err != nil {
			fmt.Println("❌ Erro ao atualizar tipo de produto:", err)
			return c.String(
				http.StatusInternalServerError,
				"Erro ao atualizar valor do tipo de produto",
			)
		}
		fmt.Println("✅ TypeProduct atualizado com sucesso:", tp)

		// Redirecionamento
		c.Response().Header().Set("HX-Redirect", "../")
		fmt.Println("✅ Finalizado com sucesso")
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
