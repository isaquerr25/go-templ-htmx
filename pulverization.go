package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/isaquerr25/go-templ-htmx/views/pages/pulverization"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func ListPulverizations(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var pulverizations []Pulverization

		if err := db.Find(&pulverizations).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao buscar pulverizações")
		}

		var items []pulverization.PulverizationItem
		for _, p := range pulverizations {
			items = append(items, pulverization.PulverizationItem{
				ID:           p.ID,
				PlantingID:   p.PlantingID,
				ProductID:    p.ProductID,
				AppliedAt:    p.AppliedAt.Format("2006-01-02"),
				QuantityUsed: p.QuantityUsed,
				Unit:         p.Unit,
			})
		}

		return pulverization.List(items).Render(c.Request().Context(), c.Response().Writer)
	}
}

func CreatePulverization(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var p pulverization.PulverizationProps
		if err := c.Bind(&p); err != nil {
			return c.String(http.StatusBadRequest, "Erro ao ler dados do formulário")
		}

		appliedAt, err := time.Parse("2006-01-02", p.AppliedAt)
		if err != nil {
			p.Error = map[string]string{"AppliedAt": "Data inválida"}
			return pulverization.Index(p).
				Render(c.Request().Context(), c.Response().Writer)

		}

		newPulverization := Pulverization{
			PlantingID:   p.PlantingID,
			ProductID:    p.ProductID,
			AppliedAt:    appliedAt,
			QuantityUsed: p.QuantityUsed,
			Unit:         p.Unit,
		}

		if err := db.Create(&newPulverization).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao criar pulverização")
		}

		return ListPulverizations(db)(c)
	}
}

func UpdatePulverization(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))

		var p pulverization.PulverizationProps
		if err := c.Bind(&p); err != nil {
			return c.String(http.StatusBadRequest, "Erro ao ler dados do formulário")
		}

		appliedAt, err := time.Parse("2006-01-02", p.AppliedAt)
		if err != nil {
			p.Error = map[string]string{"AppliedAt": "Data inválida"}
			return pulverization.Index(p).
				Render(c.Request().Context(), c.Response().Writer)
		}

		var pul Pulverization
		if err := db.First(&pul, id).Error; err != nil {
			return c.String(http.StatusNotFound, "Pulverização não encontrada")
		}

		pul.PlantingID = p.PlantingID
		pul.ProductID = p.ProductID
		pul.AppliedAt = appliedAt
		pul.QuantityUsed = p.QuantityUsed
		pul.Unit = p.Unit

		if err := db.Save(&pul).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao atualizar pulverização")
		}

		return ListPulverizations(db)(c)
	}
}

func DeletePulverization(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))

		if err := db.Delete(&Pulverization{}, id).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao deletar pulverização")
		}

		return ListPulverizations(db)(c)
	}
}

func ShowPulverizationForm(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		if id == "" {
			return pulverization.Index(pulverization.PulverizationProps{}).
				Render(c.Request().Context(), c.Response().Writer)
		}

		var pul Pulverization
		if err := db.First(&pul, id).Error; err != nil {
			return c.String(http.StatusNotFound, "Pulverização não encontrada")
		}

		p := pulverization.PulverizationProps{
			ID:           pul.ID,
			PlantingID:   pul.PlantingID,
			ProductID:    pul.ProductID,
			AppliedAt:    pul.AppliedAt.Format("2006-01-02"),
			QuantityUsed: pul.QuantityUsed,
			Unit:         pul.Unit,
		}

		return pulverization.Index(p).Render(c.Request().Context(), c.Response().Writer)
	}
}
