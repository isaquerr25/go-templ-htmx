package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/isaquerr25/go-templ-htmx/views/pages/planting"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func ListPlantings(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var plantings []Planting

		if err := db.Find(&plantings).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao buscar plantios")
		}

		var items []planting.PlantingItem
		for _, p := range plantings {
			items = append(items, planting.PlantingItem{
				ID:          p.ID,
				FieldID:     p.FieldID,
				CropName:    p.CropName,
				StartedAt:   p.StartedAt,
				EndedAt:     p.EndedAt,
				IsCompleted: p.IsCompleted,
				AreaUsed:    p.AreaUsed,
			})
		}

		// Gerar HTML via templ do go-templ-htmx
		return planting.List(items).Render(c.Request().Context(), c.Response().Writer)
		// Responder com HTML
	}
}

func CreatePlanting(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var p planting.PlantingProps
		if err := c.Bind(&p); err != nil {
			return c.String(http.StatusBadRequest, "Erro ao ler dados do formulário")
		}

		startedAt, err := time.Parse("2006-01-02", p.StartedAt)
		if err != nil {
			p.Error = map[string]string{"StartedAt": "Data inválida"}
			return c.Render(http.StatusOK, "main", planting.Index(p))
		}

		var endedAt *time.Time
		if p.EndedAt != "" {
			t, err := time.Parse("2006-01-02", p.EndedAt)
			if err != nil {
				p.Error = map[string]string{"EndedAt": "Data final inválida"}
				return c.Render(http.StatusOK, "main", planting.Index(p))
			}
			endedAt = &t
		}

		newPlanting := Planting{
			FieldID:     p.FieldID,
			CropName:    p.CropName,
			StartedAt:   startedAt,
			EndedAt:     endedAt,
			IsCompleted: p.IsCompleted,
			AreaUsed:    p.AreaUsed,
		}

		if err := db.Create(&newPlanting).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao criar plantio")
		}

		return ListPlantings(db)(c)
	}
}

func UpdatePlanting(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))

		var p planting.PlantingProps
		if err := c.Bind(&p); err != nil {
			return c.String(http.StatusBadRequest, "Erro ao ler dados do formulário")
		}

		startedAt, err := time.Parse("2006-01-02", p.StartedAt)
		if err != nil {
			p.Error = map[string]string{"StartedAt": "Data inválida"}

			return planting.Index(p).
				Render(c.Request().Context(), c.Response().Writer)

		}

		var endedAt *time.Time
		if p.EndedAt != "" {
			t, err := time.Parse("2006-01-02", p.EndedAt)
			if err != nil {
				p.Error = map[string]string{"EndedAt": "Data final inválida"}

				return planting.Index(p).
					Render(c.Request().Context(), c.Response().Writer)
			}
			endedAt = &t
		}

		var plant Planting
		if err := db.First(&plant, id).Error; err != nil {
			return c.String(http.StatusNotFound, "Plantio não encontrado")
		}

		plant.FieldID = p.FieldID
		plant.CropName = p.CropName
		plant.StartedAt = startedAt
		plant.EndedAt = endedAt
		plant.IsCompleted = p.IsCompleted
		plant.AreaUsed = p.AreaUsed

		if err := db.Save(&plant).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao atualizar plantio")
		}

		return ListPlantings(db)(c)
	}
}

func DeletePlanting(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))

		if err := db.Delete(&Planting{}, id).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao deletar plantio")
		}

		return ListPlantings(db)(c)
	}
}

func ShowPlantingForm(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		if id == "" {
			return planting.Index(planting.PlantingProps{}).
				Render(c.Request().Context(), c.Response().Writer)
		}

		var plant Planting
		if err := db.First(&plant, id).Error; err != nil {
			return c.String(http.StatusNotFound, "Plantio não encontrado")
		}

		var endedAt string
		if plant.EndedAt != nil {
			endedAt = plant.EndedAt.Format("2006-01-02")
		}

		p := planting.PlantingProps{
			ID:          plant.ID,
			FieldID:     plant.FieldID,
			CropName:    plant.CropName,
			StartedAt:   plant.StartedAt.Format("2006-01-02"),
			EndedAt:     endedAt,
			IsCompleted: plant.IsCompleted,
			AreaUsed:    plant.AreaUsed,
		}

		// Gerar HTML via templ do go-templ-htmx
		return planting.Index(p).Render(c.Request().Context(), c.Response().Writer)
	}
}
