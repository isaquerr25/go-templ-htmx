package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/isaquerr25/go-templ-htmx/views/pages/pulverization"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func ListPulverizations() echo.HandlerFunc {
	return func(c echo.Context) error {
		var pulverizations []Pulverization

		// Busca só as pulverizações
		if err := db.Find(&pulverizations).Error; err != nil {
			fmt.Printf("Erro ao buscar pulverizações: %v\n", err)
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				"Erro interno ao buscar pulverizações",
			)
		}

		var items []pulverization.PulverizationProps
		for _, p := range pulverizations {
			var products []pulverization.ProductInput

			// Busca os produtos aplicados manualmente para cada pulverização
			if err := db.Debug().Find(&pulverizations).Error; err != nil {
				fmt.Printf("Erro ao buscar pulverizações: %v\n", err)
				return echo.NewHTTPError(
					http.StatusInternalServerError,
					"Erro interno ao buscar pulverizações",
				)
			}

			var prods []pulverization.ProductInput
			for _, ap := range products {
				prods = append(prods, pulverization.ProductInput{
					ProductID:    ap.ProductID,
					QuantityUsed: ap.QuantityUsed,
				})
			}

			items = append(items, pulverization.PulverizationProps{
				ID:         p.ID,
				PlantingID: p.PlantingID,
				AppliedAt: pulverization.Date{
					Time: p.AppliedAt,
				},
				Unit:     p.Unit,
				Products: prods,
			})
		}

		if err := pulverization.List(items).Render(c.Request().Context(), c.Response().Writer); err != nil {
			fmt.Printf("Erro ao renderizar a lista de pulverizações: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Erro ao renderizar resposta")
		}

		return nil
	}
}

func CreatePulverization(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		planId := c.Param("planId")
		planIdInt, err := strconv.Atoi(planId)
		if err != nil {
			return c.String(http.StatusBadRequest, "planId inválido")
		}

		// Extrair campos individuais
		appliedAtStr := c.FormValue("appliedAt")
		unit := c.FormValue("unit")

		appliedAt, err := time.Parse("2006-01-02", appliedAtStr)
		if err != nil {
			return c.String(http.StatusBadRequest, "Data (AppliedAt) inválida")
		}

		// Criar a pulverização principal
		newPulverization := Pulverization{
			PlantingID: uint(planIdInt),
			AppliedAt:  appliedAt,
			Unit:       unit,
		}

		if err := db.Create(&newPulverization).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao criar pulverização")
		}

		// Processar os produtos aplicados
		form := c.Request().PostForm
		productIDs := form["Products[].ProductID"]
		quantities := form["Products[].QuantityUsed"]

		for i := range productIDs {
			productID, err := strconv.Atoi(productIDs[i])
			if err != nil {
				continue // Ignora produto inválido
			}

			quantity, err := strconv.ParseFloat(quantities[i], 64)
			if err != nil {
				continue // Ignora quantidade inválida
			}

			applied := AppliedProduct{
				PulverizationID: newPulverization.ID,
				ProductID:       uint(productID),
				QuantityUsed:    quantity,
			}

			if err := db.Create(&applied).Error; err != nil {
				return c.String(http.StatusInternalServerError, "Erro ao salvar produto aplicado")
			}
		}

		c.Response().Header().Set("HX-Redirect", "../")
		return c.String(http.StatusOK, "")
	}
}

func UpdatePulverization(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))

		var p pulverization.PulverizationProps
		if err := c.Bind(&p); err != nil {
			return c.String(http.StatusBadRequest, "Erro ao ler dados do formulário")
		}

		var pul Pulverization
		if err := db.Preload("Products").First(&pul, id).Error; err != nil {
			return c.String(http.StatusNotFound, "Pulverização não encontrada")
		}

		// Atualiza os campos principais
		pul.PlantingID = p.PlantingID
		pul.AppliedAt = p.AppliedAt.Time
		pul.Unit = p.Unit

		// Atualiza no banco
		if err := db.Save(&pul).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao atualizar pulverização")
		}

		// Remove produtos antigos
		if err := db.Where("pulverization_id = ?", pul.ID).Delete(&AppliedProduct{}).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao limpar produtos antigos")
		}

		// Adiciona novos produtos
		for _, prod := range p.Products {
			applied := AppliedProduct{
				PulverizationID: pul.ID,
				ProductID:       prod.ProductID,
				QuantityUsed:    prod.QuantityUsed,
			}
			if err := db.Create(&applied).Error; err != nil {
				return c.String(http.StatusInternalServerError, "Erro ao salvar produtos")
			}
		}

		return ListPulverizations()(c)
	}
}

func DeletePulverization(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))

		if err := db.Delete(&Pulverization{}, id).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao deletar pulverização")
		}

		return ListPulverizations()(c)
	}
}

func ShowPulverizationForm(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")

		a, _ := GetAllProductsProps()
		b, _ := GetAllPlantings()
		if id == "" {
			// Se for criação, renderiza formulário vazio
			return pulverization.Index(pulverization.PulverizationProps{
				ID:         0,
				PlantingID: 0,
				Unit:       "",
				Products:   []pulverization.ProductInput{},
				Error:      map[string]string{},
				AppliedAt: pulverization.Date{
					Time: time.Now(),
				},
			}, pulverization.UseProps{
				Prod: a,
				Plan: b,
			}).
				Render(c.Request().Context(), c.Response().Writer)
		}

		// Busca a pulverização pelo ID
		var pul Pulverization
		if err := db.Preload("Products").First(&pul, id).Error; err != nil {
			return c.String(http.StatusNotFound, "Pulverização não encontrada")
		}

		// Mapeia os produtos para o formato do front
		var products []pulverization.ProductInput
		for _, prod := range pul.Products {
			products = append(products, pulverization.ProductInput{
				ProductID:    prod.ProductID,
				QuantityUsed: prod.QuantityUsed,
			})
		}

		p := pulverization.PulverizationProps{
			ID:         pul.ID,
			PlantingID: pul.PlantingID,
			AppliedAt: pulverization.Date{
				Time: pul.AppliedAt,
			},
			Unit:     pul.Unit,
			Products: products,
		}

		return pulverization.Index(p, pulverization.UseProps{
			Prod: a,
			Plan: b,
		}).
			Render(c.Request().Context(), c.Response().Writer)
	}
}
