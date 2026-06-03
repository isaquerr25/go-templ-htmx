package main

import (
	"fmt"
	"math"
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
		if err := db.Find(&pulverizations).Error; err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Erro ao buscar pulverizações")
		}

		var items []pulverization.PulverizationProps
		for _, p := range pulverizations {
			var products []pulverization.ProductInput
			var prods []AppliedProduct
			db.Where("pulverization_id = ?", p.ID).Find(&prods)
			for _, ap := range prods {
				products = append(products, pulverization.ProductInput{
					ProductID:    ap.ProductID,
					QuantityUsed: ap.QuantityUsed,
				})
			}

			items = append(items, pulverization.PulverizationProps{
				ID:         p.ID,
				PlantingID: p.PlantingID,
				AppliedAt:  pulverization.Date{Time: p.AppliedAt},
				Unit:       p.Unit,
				Products:   products,
			})
		}

		return pulverization.List(items).Render(c.Request().Context(), c.Response().Writer)
	}
}

func UpdatePulverization(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))
		var p pulverization.PulverizationProps
		if err := c.Bind(&p); err != nil {
			return c.String(http.StatusBadRequest, "Erro ao ler dados")
		}

		var pul Pulverization
		if err := db.Preload("Products").First(&pul, id).Error; err != nil {
			return c.String(http.StatusNotFound, "Pulverização não encontrada")
		}

		_ = db.Transaction(func(tx *gorm.DB) error {
			// Restore old stock
			var oldProds []AppliedProduct
			tx.Where("pulverization_id = ?", pul.ID).Find(&oldProds)
			for _, ap := range oldProds {
				restoreStock(tx, ap.ProductID, ap.QuantityUsed)
			}
			tx.Where("pulverization_id = ?", pul.ID).Delete(&AppliedProduct{})

			pul.PlantingID = p.PlantingID
			pul.AppliedAt = p.AppliedAt.Time
			pul.Unit = p.Unit
			tx.Save(&pul)

			for _, prod := range p.Products {
				if err := consumeStock(tx, prod.ProductID, prod.QuantityUsed); err != nil {
					return err
				}
				applied := AppliedProduct{
					PulverizationID: pul.ID,
					ProductID:       prod.ProductID,
					QuantityUsed:    prod.QuantityUsed,
				}
				tx.Create(&applied)
			}
			return nil
		})

		return ListPulverizations()(c)
	}
}

func DeletePulverization(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))
		db.Transaction(func(tx *gorm.DB) error {
			var prods []AppliedProduct
			tx.Where("pulverization_id = ?", id).Find(&prods)
			for _, ap := range prods {
				restoreStock(tx, ap.ProductID, ap.QuantityUsed)
			}
			tx.Delete(&Pulverization{}, id)
			return nil
		})
		c.Response().Header().Set("HX-Refresh", "true")
		return c.NoContent(http.StatusOK)
	}
}

func ShowPulverizationForm(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		planIDStr := c.Param("planId")

		a, _ := GetAllProductsProps()
		b, _ := GetAllPlantings()

		if id == "" {
			props := pulverization.PulverizationProps{
				ID:         0,
				PlantingID: 0,
				Unit:       "",
				Products:   []pulverization.ProductInput{},
				Error:      map[string]string{},
				AppliedAt:  pulverization.Date{Time: time.Now()},
			}
			if planIDStr != "" {
				if pid, err := strconv.Atoi(planIDStr); err == nil {
					props.PlantingID = uint(pid)
				}
			}
			return pulverization.Index(props, pulverization.UseProps{Prod: a, Plan: b}).
				Render(c.Request().Context(), c.Response().Writer)
		}

		var pul Pulverization
		if err := db.Preload("Products").First(&pul, id).Error; err != nil {
			return c.String(http.StatusNotFound, "Pulverização não encontrada")
		}

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
			AppliedAt:  pulverization.Date{Time: pul.AppliedAt},
			Unit:       pul.Unit,
			Products:   products,
		}

		return pulverization.Index(p, pulverization.UseProps{Prod: a, Plan: b}).
			Render(c.Request().Context(), c.Response().Writer)
	}
}

func CreatePulverization(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		planIdStr := c.Param("planId")
		planIdInt, err := strconv.Atoi(planIdStr)
		if err != nil {
			return c.String(http.StatusBadRequest, "planId inválido")
		}

		appliedAtStr := c.FormValue("appliedAt")
		appliedAt, err := time.Parse("2006-01-02", appliedAtStr)
		if err != nil {
			return c.String(http.StatusBadRequest, "Data inválida")
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			pulv := Pulverization{
				PlantingID: uint(planIdInt),
				AppliedAt:  appliedAt,
				Products:   []AppliedProduct{},
			}
			if err := tx.Create(&pulv).Error; err != nil {
				return err
			}

			form := c.Request().PostForm
			validProducts := 0
			i := 0
			for {
				keyID := fmt.Sprintf("products[%d].productId", i)
				keyQty := fmt.Sprintf("products[%d].quantityUsed", i)
				idStr := form.Get(keyID)
				qtyStr := form.Get(keyQty)

				if idStr == "" && qtyStr == "" {
					break
				}

				productID, err := strconv.Atoi(idStr)
				if err != nil {
					i++
					continue
				}
				rawQuantity, err := strconv.ParseFloat(qtyStr, 64)
				if err != nil {
					i++
					continue
				}
				quantity := rawQuantity / 1000

				if err := consumeStock(tx, uint(productID), quantity); err != nil {
					var prod Product
					tx.First(&prod, productID)
					return fmt.Errorf("estoque insuficiente para %s", prod.Name)
				}

				applied := AppliedProduct{
					PulverizationID: pulv.ID,
					ProductID:       uint(productID),
					QuantityUsed:    quantity,
				}
				tx.Create(&applied)
				validProducts++
				i++
			}
			if validProducts == 0 {
				return fmt.Errorf("deve conter no mínimo 1 produto válido")
			}
			return nil
		})
		if err != nil {
			p := pulverization.PulverizationProps{
				Error: map[string]string{"Form": err.Error()},
			}
			a, _ := GetAllProductsProps()
			b, _ := GetAllPlantings()
			return pulverization.Index(p, pulverization.UseProps{Prod: a, Plan: b}).
				Render(c.Request().Context(), c.Response().Writer)
		}

		c.Response().Header().Set("HX-Redirect", "../")
		return c.String(http.StatusOK, "")
	}
}

func CreatePulverizationWithSplit(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		form := c.Request().PostForm
		if form == nil {
			c.Request().ParseForm()
			form = c.Request().PostForm
		}

		appliedAtStr := form.Get("appliedAt")
		unit := form.Get("unit")
		appliedAt, err := time.Parse("2006-01-02", appliedAtStr)
		if err != nil {
			p := pulverization.PulverizationProps{
				Error: map[string]string{"AppliedAt": "Data inválida"},
			}
			a, _ := GetAllProductsProps()
			b, _ := GetAllPlantings()
			return pulverization.Index(p, pulverization.UseProps{Prod: a, Plan: b}).
				Render(c.Request().Context(), c.Response().Writer)
		}

		type prodTotal struct {
			ProductID uint
			Quantity  float64
		}
		var totals []prodTotal
		pIdx := 0
		for {
			keyID := fmt.Sprintf("products[%d].productId", pIdx)
			keyQty := fmt.Sprintf("products[%d].quantityUsed", pIdx)
			idStr := form.Get(keyID)
			qtyStr := form.Get(keyQty)
			if idStr == "" && qtyStr == "" {
				break
			}
			pIdx++
			idInt, err := strconv.Atoi(idStr)
			if err != nil {
				continue
			}
			rawQty, err := strconv.ParseFloat(qtyStr, 64)
			if err != nil {
				continue
			}
			quantity := rawQty / 1000.0
			totals = append(totals, prodTotal{ProductID: uint(idInt), Quantity: quantity})
		}

		if len(totals) == 0 {
			p := pulverization.PulverizationProps{
				Error: map[string]string{"Form": "Informe ao menos 1 produto"},
			}
			a, _ := GetAllProductsProps()
			b, _ := GetAllPlantings()
			return pulverization.Index(p, pulverization.UseProps{Prod: a, Plan: b}).
				Render(c.Request().Context(), c.Response().Writer)
		}

		type subdiv struct {
			PlantingID uint
			Percent    float64
		}
		var subs []subdiv
		sumPercent := 0.0
		for sIdx := 0; ; sIdx++ {
			keyPlant := fmt.Sprintf("plans[%d].planId", sIdx)
			keyPerc := fmt.Sprintf("plans[%d].quantityUsed", sIdx)
			plantStr := form.Get(keyPlant)
			percStr := form.Get(keyPerc)
			if plantStr == "" && percStr == "" {
				break
			}
			plantInt, _ := strconv.Atoi(plantStr)
			perc, _ := strconv.ParseFloat(percStr, 64)
			sumPercent += perc
			subs = append(subs, subdiv{PlantingID: uint(plantInt), Percent: perc})
		}

		if len(subs) == 0 {
			p := pulverization.PulverizationProps{
				Error: map[string]string{"Form": "Informe ao menos 1 plantio"},
			}
			a, _ := GetAllProductsProps()
			b, _ := GetAllPlantings()
			return pulverization.Index(p, pulverization.UseProps{Prod: a, Plan: b}).
				Render(c.Request().Context(), c.Response().Writer)
		}

		if math.Abs(sumPercent-100.0) > 0.0001 {
			p := pulverization.PulverizationProps{
				Error: map[string]string{
					"Form": fmt.Sprintf("Soma das porcentagens deve ser 100 (atual %.2f)", sumPercent),
				},
			}
			a, _ := GetAllProductsProps()
			b, _ := GetAllPlantings()
			return pulverization.Index(p, pulverization.UseProps{Prod: a, Plan: b}).
				Render(c.Request().Context(), c.Response().Writer)
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			for _, t := range totals {
				if err := consumeStock(tx, t.ProductID, t.Quantity); err != nil {
					return fmt.Errorf("estoque insuficiente para produto %d", t.ProductID)
				}
			}
			for _, s := range subs {
				pulv := Pulverization{
					PlantingID: s.PlantingID,
					AppliedAt:  appliedAt,
					Unit:       unit,
					Products:   []AppliedProduct{},
				}
				if err := tx.Create(&pulv).Error; err != nil {
					return err
				}
				for _, t := range totals {
					share := t.Quantity * (s.Percent / 100.0)
					applied := AppliedProduct{
						PulverizationID: pulv.ID,
						ProductID:       t.ProductID,
						QuantityUsed:    share,
					}
					tx.Create(&applied)
				}
			}
			return nil
		})
		if err != nil {
			p := pulverization.PulverizationProps{
				Error: map[string]string{"Form": err.Error()},
			}
			a, _ := GetAllProductsProps()
			b, _ := GetAllPlantings()
			return pulverization.Index(p, pulverization.UseProps{Prod: a, Plan: b}).
				Render(c.Request().Context(), c.Response().Writer)
		}

		c.Response().Header().Set("HX-Redirect", "../")
		return c.String(http.StatusOK, "")
	}
}

// CreatePulverizationTotalArea distribui produtos por todos os plantios ativos proporcional à área
func CreatePulverizationTotalArea(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Request().ParseForm()
		form := c.Request().PostForm

		appliedAtStr := form.Get("appliedAt")
		appliedAt, err := time.Parse("2006-01-02", appliedAtStr)
		if err != nil {
			return c.String(http.StatusBadRequest, "Data inválida")
		}

		type prodTotal struct {
			ProductID uint
			Quantity  float64
		}
		var totals []prodTotal
		pIdx := 0
		for {
			keyID := fmt.Sprintf("products[%d].productId", pIdx)
			keyQty := fmt.Sprintf("products[%d].quantityUsed", pIdx)
			idStr := form.Get(keyID)
			qtyStr := form.Get(keyQty)
			if idStr == "" && qtyStr == "" {
				break
			}
			pIdx++
			idInt, _ := strconv.Atoi(idStr)
			rawQty, _ := strconv.ParseFloat(qtyStr, 64)
			quantity := rawQty / 1000.0
			totals = append(totals, prodTotal{ProductID: uint(idInt), Quantity: quantity})
		}
		if len(totals) == 0 {
			return c.String(http.StatusBadRequest, "Informe ao menos 1 produto")
		}

		// Buscar plantios ativos (não completos, não em morte)
		var plantings []Planting
		db.Where("is_completed = ?", false).Find(&plantings)

		// Filtrar plantios em estágio ativo (excluir morte)
		var active []Planting
		var totalArea float64
		for _, pl := range plantings {
			if pl.TypeProductID != nil {
				var tp TypeProduct
				if err := db.First(&tp, *pl.TypeProductID).Error; err == nil {
					stage := computeCurrentStage(pl.StartedAt, tp)
					if stage == "morte" {
						continue
					}
				}
			}
			active = append(active, pl)
			totalArea += pl.AreaUsed
		}

		if len(active) == 0 || totalArea <= 0 {
			return c.String(http.StatusBadRequest, "Nenhum plantio ativo disponível")
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			for _, t := range totals {
				if err := consumeStock(tx, t.ProductID, t.Quantity); err != nil {
					return fmt.Errorf("estoque insuficiente para produto %d", t.ProductID)
				}
			}

			for _, pl := range active {
				share := pl.AreaUsed / totalArea
				pulv := Pulverization{
					PlantingID: pl.ID,
					AppliedAt:  appliedAt,
					Products:   []AppliedProduct{},
				}
				tx.Create(&pulv)
				for _, t := range totals {
					qty := t.Quantity * share
					applied := AppliedProduct{
						PulverizationID: pulv.ID,
						ProductID:       t.ProductID,
						QuantityUsed:    qty,
					}
					tx.Create(&applied)
				}
			}
			return nil
		})
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		c.Response().Header().Set("HX-Redirect", "/")
		return c.String(http.StatusOK, "")
	}
}
