package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/isaquerr25/go-templ-htmx/views/pages/service"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func CreateService(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		fmt.Println("👉 Início da função CreateService")

		planId := c.Param("planId")

		req := c.Request()
		if err := req.ParseForm(); err != nil {
			fmt.Println("❌ Erro ao fazer ParseForm:", err)
			return c.String(http.StatusBadRequest, "Erro ao processar formulário")
		}

		fmt.Println("📥 Dados do formulário:", req.Form.Encode())

		planIdInt, err := strconv.Atoi(planId)
		if err != nil {
			fmt.Println("Erro ao converter planId:", err) // ✅ 3
			return c.String(http.StatusBadRequest, "planId inválido")
		}

		var s Service

		// Leitura dos campos do formulário
		plantingID := uint(planIdInt)
		s.PlantingID = &plantingID
		s.Name = c.FormValue("name")
		s.Description = c.FormValue("description")
		s.Notes = c.FormValue("notes")

		fmt.Println("📌 Name:", s.Name)
		fmt.Println("📝 Description:", s.Description)
		fmt.Println("🗒️ Notes:", s.Notes)

		// Custo
		if costStr := c.FormValue("cost"); costStr != "" {
			fmt.Println("💰 Cost (string):", costStr)
			if _, err := fmt.Sscanf(costStr, "%f", &s.Cost); err != nil {
				fmt.Println("❌ Erro ao converter cost:", err)
			}
		}

		// PlantingID
		if plantingIdStr := c.FormValue("plantingId"); plantingIdStr != "" {
			fmt.Println("🌱 PlantingID (string):", plantingIdStr)
			var plantingId uint
			if _, err := fmt.Sscanf(plantingIdStr, "%d", &plantingId); err != nil {
				fmt.Println("❌ Erro ao converter plantingId:", err)
			} else {
				s.PlantingID = &plantingId
				fmt.Println("✅ PlantingID (uint):", *s.PlantingID)
			}
		}

		// Data
		if dateStr := c.FormValue("performedAt"); dateStr != "" {
			fmt.Println("📅 performedAt (string):", dateStr)
			parsedDate, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				fmt.Println("❌ Erro ao converter data:", err)
			} else {
				s.CreateAt = parsedDate
				fmt.Println("✅ Data convertida:", s.CreateAt)
			}
		}

		// Tentativa de salvar no banco
		fmt.Println("🚀 Salvando serviço no banco:", s)

		if err := db.Create(&s).Error; err != nil {
			fmt.Println("❌ Erro ao salvar no banco:", err)
			return c.String(http.StatusInternalServerError, "Erro ao salvar serviço")
		}

		fmt.Println("✅ Serviço salvo com sucesso. Redirecionando.")

		c.Response().Header().Set("HX-Redirect", "./")
		return c.String(http.StatusOK, "")
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

		c.Response().Header().Set("HX-Redirect", "../")
		return c.String(http.StatusOK, "")
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
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			return c.String(http.StatusBadRequest, "ID inválido")
		}

		if err := db.Delete(&Service{}, id).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao excluir serviço")
		}

		c.Response().Header().Set("HX-Redirect", "")
		return c.String(http.StatusOK, "")
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
