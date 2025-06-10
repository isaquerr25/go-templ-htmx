package main

import (
	"net/http"
	"strconv"

	"github.com/isaquerr25/go-templ-htmx/views/pages/fertilization"
	"github.com/isaquerr25/go-templ-htmx/views/pages/pulverization"
	"github.com/labstack/echo/v4"
)

func ListFertilization(c echo.Context) error {
	var ferts []Fertilization
	if err := db.Find(&ferts).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Erro ao buscar fertilizações")
	}

	var items []fertilization.FertilizationProps
	for _, f := range ferts {
		var products []fertilization.ApplyFertilizationProps
		for _, p := range f.Products {
			products = append(products, fertilization.ApplyFertilizationProps{
				ProductID:    p.ProductID,
				QuantityUsed: p.QuantityUsed,
				Unit:         p.Unit,
			})
		}

		items = append(items, fertilization.FertilizationProps{
			ID:              f.ID,
			PlantingID:      f.PlantingID,
			ApplicationType: f.ApplicationType,
			AppliedAt:       f.AppliedAt,
			Products:        []pulverization.ProductInput{},
		})
	}

	return fertilization.List(fertilization.FertilizationListProps{Items: items}).
		Render(c.Request().Context(), c.Response())
}

func ShowFertilization(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	var f Fertilization
	if err := db.Preload("Products").First(&f, id).Error; err != nil {
		return c.String(http.StatusNotFound, "Fertilização não encontrada")
	}

	var products []fertilization.ApplyFertilizationProps
	for _, p := range f.Products {
		products = append(products, fertilization.ApplyFertilizationProps{
			ProductID:    p.ProductID,
			QuantityUsed: p.QuantityUsed,
			Unit:         p.Unit,
		})
	}

	p := fertilization.FertilizationProps{
		ID:              f.ID,
		PlantingID:      f.PlantingID,
		ApplicationType: f.ApplicationType,
		AppliedAt:       f.AppliedAt,
		Products:        []pulverization.ProductInput{},
	}

	return fertilization.Index(p, pulverization.UseProps{}).
		Render(c.Request().Context(), c.Response())
}

func CreateFertilization(c echo.Context) error {
	// Supondo que você receba os dados dos produtos via JSON ou formulário com repetição (ex: products[0].productId, products[0].quantityUsed...)

	var input fertilization.FertilizationProps
	if err := c.Bind(&input); err != nil {
		return c.String(http.StatusBadRequest, "Erro ao fazer bind dos dados do formulário")
	}

	input.Error = make(map[string]string)

	// Validação simplificada, adaptar conforme necessário
	if input.PlantingID == 0 {
		input.Error["PlantingID"] = "ID de plantio inválido"
	}
	if input.ApplicationType == "" {
		input.Error["ApplicationType"] = "Tipo de aplicação obrigatório"
	}
	if input.AppliedAt.IsZero() {
		input.Error["AppliedAt"] = "Data inválida ou não informada"
	}

	// Validar os produtos
	if len(input.Products) == 0 {
		input.Error["Products"] = "É necessário pelo menos um produto aplicado"
	} else {
		for i, p := range input.Products {
			if p.ProductID == 0 {
				input.Error["Products"] = "Produto inválido no item " + strconv.Itoa(i+1)
			}
			if p.QuantityUsed <= 0 {
				input.Error["Products"] = "Quantidade usada inválida no item " + strconv.Itoa(i+1)
			}

		}
	}

	if len(input.Error) > 0 {
		return fertilization.Index(input, pulverization.UseProps{}).
			Render(c.Request().Context(), c.Response())
	}

	// Criar registro principal
	f := Fertilization{
		PlantingID:      input.PlantingID,
		ApplicationType: input.ApplicationType,
		AppliedAt:       input.AppliedAt,
	}

	if err := db.Create(&f).Error; err != nil {
		input.Error["global"] = "Erro ao salvar fertilização"
		return fertilization.Index(input, pulverization.UseProps{}).
			Render(c.Request().Context(), c.Response())
	}

	// Criar os ApplyFertilization vinculados
	for _, p := range input.Products {
		af := ApplyFertilization{
			FertilizationID: f.ID,
			ProductID:       p.ProductID,
			QuantityUsed:    p.QuantityUsed,
		}
		if err := db.Create(&af).Error; err != nil {
			input.Error["global"] = "Erro ao salvar produtos aplicados"
			return fertilization.Index(input, pulverization.UseProps{}).
				Render(c.Request().Context(), c.Response())
		}
	}

	return ListFertilization(c)
}

func UpdateFertilization(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	var f Fertilization
	if err := db.Preload("Products").First(&f, id).Error; err != nil {
		return c.String(http.StatusNotFound, "Fertilização não encontrada")
	}

	var input fertilization.FertilizationProps
	if err := c.Bind(&input); err != nil {
		return c.String(http.StatusBadRequest, "Erro no bind")
	}

	// Atualizar os campos do fertilization
	f.PlantingID = input.PlantingID
	f.ApplicationType = input.ApplicationType
	f.AppliedAt = input.AppliedAt

	if err := db.Save(&f).Error; err != nil {
		input.Error = map[string]string{"global": "Erro ao atualizar fertilização"}
		return fertilization.Index(input, pulverization.UseProps{}).
			Render(c.Request().Context(), c.Response())
	}

	// Para atualizar os produtos, o mais simples é apagar os antigos e criar novos (pode ser adaptado para update individual)
	if err := db.Where("fertilization_id = ?", f.ID).Delete(&ApplyFertilization{}).Error; err != nil {
		input.Error = map[string]string{"global": "Erro ao limpar produtos antigos"}
		return fertilization.Index(input, pulverization.UseProps{}).
			Render(c.Request().Context(), c.Response())
	}

	for _, p := range input.Products {
		af := ApplyFertilization{
			FertilizationID: f.ID,
			ProductID:       p.ProductID,
			QuantityUsed:    p.QuantityUsed,
		}
		if err := db.Create(&af).Error; err != nil {
			input.Error = map[string]string{"global": "Erro ao salvar produtos aplicados"}
			return fertilization.Index(input, pulverization.UseProps{}).
				Render(c.Request().Context(), c.Response())
		}
	}

	return ListFertilization(c)
}

func DeleteFertilization(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "ID inválido")
	}

	var f Fertilization
	if err := db.First(&f, id).Error; err != nil {
		return c.String(http.StatusNotFound, "Fertilização não encontrada")
	}

	if err := db.Delete(&f).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Erro ao deletar fertilização")
	}

	// Opcional: deletar também produtos aplicados
	db.Where("fertilization_id = ?", f.ID).Delete(&ApplyFertilization{})

	return c.NoContent(http.StatusNoContent)
}
