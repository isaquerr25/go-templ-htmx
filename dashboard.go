package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/isaquerr25/go-templ-htmx/views/pages/dashboard"
	"github.com/isaquerr25/go-templ-htmx/views/pages/harvest"
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

		var totalCost float64
		var totalHarvest float64

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

		var pulverizationIDs []uint
		for _, p := range pulverizations {
			pulverizationIDs = append(pulverizationIDs, p.ID)
		}

		var appliedProducts []AppliedProduct
		if len(pulverizationIDs) > 0 {
			err = db.Where("pulverization_id IN ?", pulverizationIDs).Find(&appliedProducts).Error
			if err != nil {
				fmt.Println("error aqui")
				return err
			}
		} // Monta custos a partir de serviços, irrigations e fertilizações, pulverizações

		var costs []dashboard.Cost

		for _, irr := range irrigations {
			costs = append(costs, dashboard.Cost{
				ID:          irr.ID,
				Description: "Irrigação",
				PlantingID:  0,
				CreatedAt:   time.Time{},
				Amount:      0,
				Quantity:    0,
				Type:        "irrigations",
			})
		}

		for _, irr := range appliedProducts {

			var pd Product

			db.First(&pd, irr.ProductID)

			r := regraDeTres(pd.Quantity, pd.TotalCost, irr.QuantityUsed)
			totalCost = totalCost + r
			costs = append(costs, dashboard.Cost{
				ID:          irr.ID,
				Description: pd.Name,
				PlantingID:  0,
				CreatedAt:   time.Time{},
				Amount:      r,
				Quantity:    irr.QuantityUsed,
				Type:        pd.Description,
			})
		}

		var costsServices []dashboard.Cost

		for _, irr := range services {
			totalCost = totalCost + irr.Cost
			costsServices = append(costsServices, dashboard.Cost{
				ID:          irr.ID,
				Description: irr.Name,
				Amount:      irr.Cost,
			})
		}

		var fertilizationsIDs []uint
		for _, p := range fertilizations {
			fertilizationsIDs = append(fertilizationsIDs, p.ID)
		}

		var adu []ApplyFertilization
		if len(fertilizationsIDs) > 0 {
			err = db.Where("fertilization_id IN ?", fertilizationsIDs).Find(&adu).Error
			if err != nil {
				fmt.Println("error aqui")
				return err
			}
		} // Monta custos a partir de serviços, irrigations e fertilizações, pulverizações

		// Fertilizantes
		var fertilizers []dashboard.Fertilizer
		for _, fert := range adu {
			var pd Product

			db.First(&pd, fert.ProductID)
			r := regraDeTres(pd.Quantity, pd.TotalCost, fert.QuantityUsed)
			totalCost = totalCost + r
			fertilizers = append(fertilizers, dashboard.Fertilizer{
				Amount: fmt.Sprintln(fert.QuantityUsed),
				Name:   pd.Name,
				Value:  r,
				ID:     float64(fert.ID),
			})
		}

		// Pode incluir pulverizações também em Fertilizers ou criar outro slice, depende do seu model
		// Aqui vou ignorar pulverizações para simplificar

		// Receitas a partir de vendas
		var revenues []dashboard.Revenue
		for _, sale := range sales {
			totalCost = totalCost + sale.TotalPrice
			revenues = append(revenues, dashboard.Revenue{
				ID:          sale.ID,
				Description: sale.Notes,
				Amount:      sale.TotalPrice,
				CreatedAt:   sale.DeletedAt.Time,
			})
		}

		var harvestProps []harvest.HarvestProps
		for _, h := range harvests {
			totalHarvest = totalHarvest + h.Quantity
			harvestProps = append(harvestProps, harvest.HarvestProps{
				ID: h.ID,
				HarvestedAt: harvest.Date{
					Time: h.HarvestedAt,
				},
				Quantity: h.Quantity,
				Unit:     h.Unit,
			})
		}

		var typeProduc TypeProduct
		db.First(&typeProduc, planting.TypeProductID)

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
			TypeProductProps: dashboard.TypeProductProps{
				ID:   typeProduc.ID,
				Name: typeProduc.Name,
			},
			Service:      costsServices,
			Harvest:      harvestProps,
			TotalCost:    float64(totalCost),
			TotalHarvest: totalHarvest,
		}

		return dashboard.Show(a).Render(c.Request().Context(), c.Response().Writer)
	}
}
