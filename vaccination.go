package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/isaquerr25/go-templ-htmx/views/pages/vaccination"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func ShowVaccinationForm(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		planID, err := strconv.Atoi(c.Param("planId"))
		if err != nil {
			return c.String(http.StatusBadRequest, "ID do plantio inválido")
		}

		var plant Planting
		if err := db.First(&plant, planID).Error; err != nil {
			return c.String(http.StatusNotFound, "Plantio não encontrado")
		}

		props := vaccination.VaccinationProps{
			PlantingID:   plant.ID,
			PlantingName: plant.CropName,
			VaccinatedAt: vaccination.Date{Time: time.Now()},
			Error:        map[string]string{},
		}

		return vaccination.Create(props).Render(c.Request().Context(), c.Response().Writer)
	}
}

func CreateVaccination(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		planID, err := strconv.Atoi(c.Param("planId"))
		if err != nil {
			return c.String(http.StatusBadRequest, "ID do plantio inválido")
		}

		vaccinatedAtStr := strings.TrimSpace(c.FormValue("vaccinatedAt"))
		productIDStr := strings.TrimSpace(c.FormValue("productId"))
		quantityStr := strings.TrimSpace(c.FormValue("quantity"))
		unit := strings.TrimSpace(c.FormValue("unit"))
		notes := strings.TrimSpace(c.FormValue("notes"))

		vaccinatedAt, err := time.Parse("2006-01-02", vaccinatedAtStr)
		if err != nil {
			return c.String(http.StatusBadRequest, "Data inválida")
		}

		productID, _ := strconv.Atoi(productIDStr)
		quantity, _ := strconv.ParseFloat(quantityStr, 64)

		v := Vaccination{
			PlantingID:   uint(planID),
			VaccinatedAt: vaccinatedAt,
			ProductID:    uint(productID),
			Quantity:     quantity,
			Unit:         unit,
			Notes:        notes,
		}

		if err := db.Create(&v).Error; err != nil {
			fmt.Println("Erro ao criar vacinação:", err)
			return c.String(http.StatusInternalServerError, "Erro ao salvar vacinação")
		}

		c.Response().Header().Set("HX-Redirect", fmt.Sprintf("/dashboard/plantings/%d/", planID))
		return c.String(http.StatusOK, "")
	}
}

func DeleteVaccination(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))

		if err := db.Delete(&Vaccination{}, id).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao deletar vacinação")
		}

		c.Response().Header().Set("HX-Refresh", "true")
		return c.NoContent(http.StatusOK)
	}
}

func ListVaccinations(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var vaccinations []Vaccination
		if err := db.Find(&vaccinations).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao buscar vacinações")
		}

		var items []vaccination.VaccinationItem
		for _, v := range vaccinations {
			var plantingName string
			var plant Planting
			if err := db.First(&plant, v.PlantingID).Error; err == nil {
				plantingName = plant.CropName
			}

			var productName string
			var prod Product
			if err := db.First(&prod, v.ProductID).Error; err == nil {
				productName = prod.Name
			}

			items = append(items, vaccination.VaccinationItem{
				ID:           v.ID,
				PlantingID:   v.PlantingID,
				PlantingName: plantingName,
				VaccinatedAt: v.VaccinatedAt,
				ProductName:  productName,
				Quantity:     v.Quantity,
				Unit:         v.Unit,
				Notes:        v.Notes,
			})
		}

		return vaccination.List(items).Render(c.Request().Context(), c.Response().Writer)
	}
}
