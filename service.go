package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/isaquerr25/go-templ-htmx/views/pages/service"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func CreateService(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var s Service

		// Leitura manual dos campos do formulário
		s.Name = c.FormValue("name")
		s.Description = c.FormValue("description")
		s.Notes = c.FormValue("notes")

		// Custo
		if costStr := c.FormValue("cost"); costStr != "" {
			fmt.Sscanf(costStr, "%f", &s.Cost)
		}

		// PlantingID
		if plantingIdStr := c.FormValue("plantingId"); plantingIdStr != "" {
			var plantingId uint
			fmt.Sscanf(plantingIdStr, "%d", &plantingId)
			s.PlantingID = &plantingId
		}

		// Data
		if dateStr := c.FormValue("performedAt"); dateStr != "" {
			parsedDate, err := time.Parse("2006-01-02", dateStr)
			if err == nil {
				s.CreateAt = parsedDate
			}
		}

		if err := db.Create(&s).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao salvar serviço")
		}

		return c.Redirect(http.StatusSeeOther, "/services")
	}
}

func UpdateService(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		var s Service

		// Buscar o serviço no banco
		if err := db.First(&s, id).Error; err != nil {
			return c.String(http.StatusNotFound, "Serviço não encontrado")
		}

		// Atualizar campos
		s.Name = c.FormValue("name")
		s.Description = c.FormValue("description")
		s.Notes = c.FormValue("notes")

		// Custo
		if costStr := c.FormValue("cost"); costStr != "" {
			fmt.Sscanf(costStr, "%f", &s.Cost)
		}

		// PlantingID
		if plantingIdStr := c.FormValue("plantingId"); plantingIdStr != "" {
			var plantingId uint
			fmt.Sscanf(plantingIdStr, "%d", &plantingId)
			s.PlantingID = &plantingId
		} else {
			s.PlantingID = nil
		}

		// Data
		if dateStr := c.FormValue("performedAt"); dateStr != "" {
			parsedDate, err := time.Parse("2006-01-02", dateStr)
			if err == nil {
				s.CreateAt = parsedDate
			}
		}

		if err := db.Save(&s).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao atualizar serviço")
		}

		return c.Redirect(http.StatusSeeOther, "/services")
	}
}

func NewService(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var s service.ServiceProps // Instância vazia
		s = service.ServiceProps{
			ID:          0,
			Name:        "",
			Description: "",
			Cost:        0,
			PlantingID:  new(uint),
			Notes:       "",
			CreateAt:    time.Now(),
		}
		return service.Index(s).Render(c.Request().Context(), c.Response().Writer)
	}
}

func DeleteService(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		if err := db.Delete(&Service{}, id).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao excluir serviço")
		}
		return c.Redirect(http.StatusSeeOther, "/services")
	}
}

func EditService(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		var s service.ServiceProps

		if err := db.First(&s, id).Error; err != nil {
			return c.String(http.StatusNotFound, "Serviço não encontrado")
		}

		// Renderiza o formulário preenchido
		return service.Index(s).Render(c.Request().Context(), c.Response().Writer)
	}
}
