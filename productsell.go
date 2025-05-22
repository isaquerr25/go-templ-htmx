package main

import (
	"net/http"
	"strconv"

	"github.com/isaquerr25/go-templ-htmx/views/pages/productsell"
	"github.com/labstack/echo/v4"
)

// GET /productsell
func ListProductSell(c echo.Context) error {
	var products []ProductSell
	db.Find(&products)

	var items []productsell.ProductSellProps
	for _, p := range products {
		items = append(items, productsell.ProductSellProps{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Unit:        p.Unit,
			Stock:       p.Stock,
		})
	}

	return productsell.List(productsell.ProductSellListProps{
		Items: items,
	}).Render(c.Request().Context(), c.Response().Writer)
}

// GET /productsell/create
func CreateViewProductSell(c echo.Context) error {
	return productsell.Index(productsell.ProductSellProps{}).
		Render(c.Request().Context(), c.Response().Writer)
}

// POST /productsell/create
func CreateProductSell(c echo.Context) error {
	var input ProductSell
	if err := c.Bind(&input); err != nil {
		return c.String(http.StatusBadRequest, "Erro ao processar dados")
	}

	// Validação simples
	errors := make(map[string]string)
	if input.Name == "" {
		errors["Name"] = "Nome é obrigatório"
	}

	if len(errors) > 0 {
		return productsell.Index(productsell.ProductSellProps{
			Name:        input.Name,
			Description: input.Description,
			Unit:        input.Unit,
			Price:       input.Price,
			Stock:       input.Stock,
			Error:       errors,
		}).Render(c.Request().Context(), c.Response().Writer)
	}

	if err := db.Create(&input).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Erro ao salvar produto")
	}

	return ListProductSell(c)
}

// GET /productsell/:id
func EditViewProductSell(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var p ProductSell
	if err := db.First(&p, id).Error; err != nil {
		return c.String(http.StatusNotFound, "Produto não encontrado")
	}

	return productsell.Index(productsell.ProductSellProps{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Unit:        p.Unit,
		Price:       p.Price,
		Stock:       p.Stock,
	}).Render(c.Request().Context(), c.Response().Writer)
}

// POST /productsell/update/:id
func UpdateProductSell(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var existing ProductSell
	if err := db.First(&existing, id).Error; err != nil {
		return c.String(http.StatusNotFound, "Produto não encontrado")
	}

	if err := c.Bind(&existing); err != nil {
		return c.String(http.StatusBadRequest, "Erro ao processar dados")
	}

	if err := db.Save(&existing).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Erro ao atualizar produto")
	}

	return ListProductSell(c)
}
