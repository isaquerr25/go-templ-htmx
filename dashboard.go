package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/isaquerr25/go-templ-htmx/views/pages/dashboard"
	"github.com/isaquerr25/go-templ-htmx/views/pages/planting"
	"github.com/labstack/echo/v4"
)

func ListDasboard() echo.HandlerFunc {
	return func(c echo.Context) error {
		var plantings []Planting

		if err := db.Find(&plantings).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao buscar plantios")
		}

		var items []planting.PlantingItem
		for _, p := range plantings {
			if p.IsCompleted {
				continue // pula os finalizados
			}
			items = append(items, planting.PlantingItem{
				ID:          p.ID,
				CropName:    p.CropName,
				StartedAt:   p.StartedAt,
				EndedAt:     p.EndedAt,
				IsCompleted: p.IsCompleted,
				AreaUsed:    p.AreaUsed,
			})
		}

		// Gerar HTML via templ do go-templ-htmx
		return dashboard.List(items).Render(c.Request().Context(), c.Response().Writer)
		// Responder com HTML
	}
}

func DashboardShowPlanting() echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil || id <= 0 {
			return c.NoContent(http.StatusBadRequest)
		}

		var planting Planting
		if err := db.First(&planting, id).Error; err != nil {
			return c.String(http.StatusNotFound, "Plantio não encontrado")
		}

		var irrigations []Irrigation
		db.Where("planting_id = ?", planting.ID).Find(&irrigations)

		var services []Service
		db.Where("planting_id = ?", planting.ID).Find(&services)

		var harvests []Harvest
		db.Where("planting_id = ?", planting.ID).Find(&harvests)

		var fertilizations []Fertilization
		db.Preload("Products").Where("planting_id = ?", planting.ID).Find(&fertilizations)

		var pulverizations []Pulverization
		db.Preload("Products").Where("planting_id = ?", planting.ID).Find(&pulverizations)

		var sales []Sale
		db.Where("planting_id = ?", planting.ID).Find(&sales) // melhor filtrar por planting_id

		// Monta custos a partir de serviços, irrigations e fertilizações, pulverizações
		var costs []dashboard.Cost
		for _, svc := range services {
			costs = append(costs, dashboard.Cost{
				ID:          svc.ID,
				Description: svc.Description,
				Amount:      svc.Cost,
				CreatedAt:   time.Time{},
			})
		}

		for _, irr := range irrigations {
			costs = append(costs, dashboard.Cost{
				ID:          irr.ID,
				Description: "Irrigação",
				PlantingID:  0,
				CreatedAt:   time.Time{},
			})
		}

		// Fertilizantes
		var fertilizers []dashboard.Fertilizer
		for _, fert := range fertilizations {
			fertilizers = append(fertilizers, dashboard.Fertilizer{
				Amount: "",
				Name:   fert.CreatedAt.String(),
			})
		}

		// Pode incluir pulverizações também em Fertilizers ou criar outro slice, depende do seu model
		// Aqui vou ignorar pulverizações para simplificar

		// Receitas a partir de vendas
		var revenues []dashboard.Revenue
		for _, sale := range sales {
			revenues = append(revenues, dashboard.Revenue{
				ID:          sale.ID,
				Description: sale.Notes,
				Amount:      sale.TotalPrice,
				CreatedAt:   sale.DeletedAt.Time,
			})
		}

		// Monta o objeto completo
		a := dashboard.PlantingDetailProps{
			ID:          planting.ID,
			CropName:    planting.CropName,
			AreaUsed:    planting.AreaUsed,
			StartedAt:   planting.StartedAt,
			EndedAt:     planting.EndedAt,
			IsCompleted: planting.IsCompleted,
			Costs:       costs,
			Revenues:    revenues,
			Fertilizers: fertilizers,
		}

		return dashboard.Show(a).Render(c.Request().Context(), c.Response().Writer)
	}
}
