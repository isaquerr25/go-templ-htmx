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
		fmt.Println("Início da função CreatePulverization") // ✅ 1

		planId := c.Param("planId")
		fmt.Println("planId recebido:", planId) // ✅ 2

		planIdInt, err := strconv.Atoi(planId)
		if err != nil {
			fmt.Println("Erro ao converter planId:", err) // ✅ 3
			return c.String(http.StatusBadRequest, "planId inválido")
		}
		fmt.Println("planId convertido para inteiro:", planIdInt) // ✅ 4

		// Supondo que você tenha parsing de data
		appliedAtStr := c.FormValue("appliedAt")
		fmt.Println("appliedAtStr recebido:", appliedAtStr) // ✅ 5

		appliedAt, err := time.Parse("2006-01-02", appliedAtStr)
		if err != nil {
			fmt.Println("Erro ao fazer parse da data:", err) // ✅ 6
			return c.String(http.StatusBadRequest, "Data inválida")
		}
		fmt.Println("Data aplicada convertida:", appliedAt) // ✅ 7

		// Aqui entra alguma transação
		err = db.Transaction(func(tx *gorm.DB) error {
			fmt.Println("✅ Dentro da transação") // ✅ 1

			pulv := Pulverization{
				PlantingID: uint(planIdInt),
				AppliedAt:  appliedAt,
				Products:   []AppliedProduct{},
			}

			fmt.Println("➡️ Criando Pulverization...")
			if err := tx.Create(&pulv).Error; err != nil {
				fmt.Println("❌ Erro ao criar Pulverization:", err) // ✅ 2
				return err
			}
			fmt.Println("✅ Pulverization criada com ID:", pulv.ID) // ✅ 3

			form := c.Request().PostForm
			fmt.Printf("🔍 Form recebido: %+v\n", form)

			i := 0
			for {
				keyID := fmt.Sprintf("products[%d].productId", i)
				keyQty := fmt.Sprintf("products[%d].quantityUsed", i)

				idStr := form.Get(keyID)
				qtyStr := form.Get(keyQty)

				if idStr == "" && qtyStr == "" {
					break
				}

				fmt.Printf("🔁 Produto %d: id=%s, qtd=%s\n", i, idStr, qtyStr)

				productID, err := strconv.Atoi(idStr)
				if err != nil {
					fmt.Println("⚠️ ProdutoID inválido:", idStr, err)
					i++
					continue
				}

				quantity, err := strconv.ParseFloat(qtyStr, 64)
				if err != nil {
					fmt.Println("⚠️ Quantidade inválida:", qtyStr, err)
					i++
					continue
				}

				// 🔎 Buscar o produto no banco
				var product Product
				if err := tx.First(&product, productID).Error; err != nil {
					fmt.Printf("❌ Produto com ID %d não encontrado\n", productID)
					return fmt.Errorf("produto %d não encontrado", productID)
				}

				fmt.Printf("📦 Produto encontrado: %+v\n", product)

				// ⚖️ Verificar se há quantidade suficiente
				if product.Remaining < quantity {
					fmt.Printf(
						"❌ Estoque insuficiente para produto ID %d. Em estoque: %.2f, solicitado: %.2f\n",
						productID,
						product.Remaining,
						quantity,
					)
					return fmt.Errorf("estoque insuficiente para produto %s", product.Name)
				}

				// ➖ Descontar do estoque
				product.Remaining -= quantity
				if err := tx.Save(&product).Error; err != nil {
					fmt.Println("❌ Erro ao atualizar estoque do produto:", err)
					return err
				}

				// ✅ Criar registro de aplicação
				applied := AppliedProduct{
					PulverizationID: pulv.ID,
					ProductID:       uint(productID),
					QuantityUsed:    quantity,
				}
				fmt.Printf("➡️ Salvando produto aplicado: %+v\n", applied)

				if err := tx.Create(&applied).Error; err != nil {
					fmt.Println("❌ Erro ao criar produto aplicado:", err)
					return err
				}

				fmt.Println("✅ Produto aplicado salvo")
				i++
			}
			fmt.Println("✅ Transação concluída com sucesso")
			return nil
		})
		if err != nil {
			fmt.Println("Erro na transação:", err)

			// Preenche o campo Error no props para mostrar no template
			p := pulverization.PulverizationProps{
				// ... preencha outros campos necessários para manter os dados do formulário
				Error: map[string]string{
					"Form": err.Error(),
				},
			}

			a, _ := GetAllProductsProps()
			b, _ := GetAllPlantings()

			// Renderiza o template passando p, para mostrar o erro na página
			return pulverization.Index(p, pulverization.UseProps{
				Prod: a,
				Plan: b,
			}).Render(c.Request().Context(), c.Response().Writer)
		}

		fmt.Println("Fim da função CreatePulverization") // ✅ 12
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
