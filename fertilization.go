package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/isaquerr25/go-templ-htmx/views/pages/fertilization"
	"github.com/isaquerr25/go-templ-htmx/views/pages/pulverization"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
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
			AppliedAt: fertilization.Date{
				Time: f.AppliedAt,
			},
			Products: []pulverization.ProductInput{},
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
		AppliedAt: fertilization.Date{
			Time: f.AppliedAt,
		},
		Products: []pulverization.ProductInput{},
	}

	return fertilization.Index(p, pulverization.UseProps{}).
		Render(c.Request().Context(), c.Response())
}

func CreateFertilization(c echo.Context) error {
	fmt.Println("👉 Início da função CreateFertilization")

	planId := c.Param("planId")
	fmt.Println("🔹 planId recebido:", planId)

	req := c.Request()
	if err := req.ParseForm(); err != nil {
		fmt.Println("❌ Erro ao fazer ParseForm:", err)
		return c.String(http.StatusBadRequest, "Erro ao processar formulário")
	}
	form := req.Form // Agora vai funcionar corretamente!
	fmt.Println("📥 Dados do formulário (ParseForm):", form)

	fmt.Println("📥 Dados do formulário:", form)

	input := fertilization.FertilizationProps{
		Error:           map[string]string{},
		ID:              0,
		PlantingID:      0,
		ApplicationType: "",
		AppliedAt:       fertilization.Date{},
		Products:        []pulverization.ProductInput{},
	}

	fmt.Println(form)
	// Extrai campos simples
	input.ApplicationType = form.Get("applicationType")
	fmt.Println("📌 ApplicationType:", input.ApplicationType)

	appliedAtStr := form.Get("appliedAt")
	appliedAt, err := time.Parse("2006-01-02", appliedAtStr)
	if err != nil {
		fmt.Println("❌ Erro ao converter appliedAt:", err)
		input.Error["AppliedAt"] = "Data inválida ou não informada"
	} else {
		fmt.Println("📆 appliedAt convertido:", appliedAt)
	}

	if input.ApplicationType == "" {
		fmt.Println("⚠️ ApplicationType não informado")
		input.Error["ApplicationType"] = "Tipo de aplicação obrigatório"
	}

	// Extrai produtos (repetidos)
	i := 0
	for {
		keyID := fmt.Sprintf("products[%d].productId", i)
		keyQty := fmt.Sprintf("products[%d].quantityUsed", i)

		idStr := form.Get(keyID)
		qtyStr := form.Get(keyQty)

		if idStr == "" && qtyStr == "" {
			break
		}

		fmt.Printf("🔸 Produto %d -> ID: %s | Qtd: %s\n", i, idStr, qtyStr)

		productID, errID := strconv.Atoi(idStr)
		qty, errQty := strconv.ParseFloat(qtyStr, 64)

		if errID != nil || productID == 0 {
			fmt.Printf("❌ Erro no ID do produto %d: %v\n", i, errID)
			input.Error["Products"] = "Produto inválido no item " + strconv.Itoa(i+1)
		}
		if errQty != nil || qty <= 0 {
			fmt.Printf("❌ Erro na quantidade do produto %d: %v\n", i, errQty)
			input.Error["Products"] = "Quantidade inválida no item " + strconv.Itoa(i+1)
		}

		input.Products = append(input.Products, pulverization.ProductInput{
			ProductID:    uint(productID),
			QuantityUsed: qty,
		})

		i++
	}

	if len(input.Error) > 0 {
		fmt.Println("⚠️ Erros de validação encontrados:", input.Error)
		return fertilization.Index(input, pulverization.UseProps{}).
			Render(c.Request().Context(), c.Response())
	}

	planIdInt, err := strconv.Atoi(planId)
	if err != nil {
		fmt.Println("❌ Erro ao converter planId para inteiro:", err)
		return err
	}

	// Transação segura com verificação de estoque via FIFO
	fmt.Println("🚀 Iniciando transação no banco")
	err = db.Transaction(func(tx *gorm.DB) error {
		f := Fertilization{
			PlantingID:      uint(planIdInt),
			ApplicationType: input.ApplicationType,
			AppliedAt:       appliedAt,
		}

		if err := tx.Create(&f).Error; err != nil {
			fmt.Println("❌ Erro ao salvar fertilização:", err)
			input.Error["global"] = "Erro ao salvar fertilização"
			return err
		}
		fmt.Println("✅ Fertilização criada com ID:", f.ID)

		for _, p := range input.Products {
			fmt.Println("🔍 Consumindo estoque do produto ID:", p.ProductID)

			if err := consumeStock(tx, p.ProductID, p.QuantityUsed); err != nil {
				input.Error["Products"] = fmt.Sprintf("Estoque insuficiente para produto %d", p.ProductID)
				return fmt.Errorf("estoque insuficiente para produto %d", p.ProductID)
			}

			af := ApplyFertilization{
				FertilizationID: f.ID,
				ProductID:       p.ProductID,
				QuantityUsed:    p.QuantityUsed,
			}
			if err := tx.Create(&af).Error; err != nil {
				fmt.Println("❌ Erro ao salvar aplicação de produto:", err)
				input.Error["global"] = "Erro ao salvar produtos aplicados"
				return err
			}

			fmt.Printf("✅ Produto %d aplicado com %.2f\n", p.ProductID, p.QuantityUsed)
		}

		return nil
	})
	if err != nil {
		fmt.Println("❌ Erro na transação:", err)
		return fertilization.Index(input, pulverization.UseProps{}).
			Render(c.Request().Context(), c.Response())
	}

	fmt.Println("✅ Fertilização finalizada com sucesso")
	c.Response().Header().Set("HX-Redirect", "../")
	return c.String(http.StatusOK, "")
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
	f.AppliedAt = input.AppliedAt.Time

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
	if err := db.Preload("Products").First(&f, id).Error; err != nil {
		return c.String(http.StatusNotFound, "Fertilização não encontrada")
	}

	db.Transaction(func(tx *gorm.DB) error {
		for _, p := range f.Products {
			restoreStock(tx, p.ProductID, p.QuantityUsed)
		}
		tx.Delete(&f)
		return nil
	})

	c.Response().Header().Set("HX-Redirect", "")
	return c.String(http.StatusOK, "")
}
