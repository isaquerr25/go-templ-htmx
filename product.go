package main

import (
	"fmt"

	"github.com/isaquerr25/go-templ-htmx/views/pages/produto"
	"github.com/labstack/echo/v4"
)

type Server struct{}

func GetAllProductsProps() ([]produto.ProductProps, error) {
	var products []Product
	if err := db.Find(&products).Error; err != nil {
		return nil, err
	}

	var productPropsList []produto.ProductProps
	for _, p := range products {
		productPropsList = append(productPropsList, produto.ProductProps{
			ID:                   p.ID,
			Name:                 p.Name,
			Quantity:             p.Quantity,
			Remaining:            p.Remaining,
			Unit:                 p.Unit,
			Date:                 p.Date,
			TotalCost:            p.TotalCost,
			Description:          p.Description,
			PrePulverizationBase: p.PrePulverizationBase,

			Error: map[string]string{},
		})
	}

	return productPropsList, nil
}

func GetAllProductsForUserProps() ([]produto.ProductProps, error) {
	var products []Product
	if err := db.Find(&products).Error; err != nil {
		return nil, err
	}

	var productPropsList []produto.ProductProps
	for _, p := range products {

		if p.Remaining <= 0 {
			continue
		}

		productPropsList = append(productPropsList, produto.ProductProps{
			ID:                   p.ID,
			Name:                 p.Name,
			Quantity:             p.Quantity,
			Remaining:            p.Remaining,
			Unit:                 p.Unit,
			Date:                 p.Date,
			TotalCost:            p.TotalCost,
			Description:          p.Description,
			PrePulverizationBase: p.PrePulverizationBase,

			Error: map[string]string{},
		})
	}

	return productPropsList, nil
}

func (s Server) UpdateProduct(c echo.Context) error {
	p := &Product{}
	r := db.First(p, c.Param("ID"))
	if r.Error != nil {
		fmt.Println(r.Error)
		return r.Error
	}

	k, hasError, err := validateProduct(c, p)
	if err != nil {
		return err
	}

	if !hasError {
		r := db.Save(&p)
		if r.Error != nil {
			fmt.Println(r.Error)

			return r.Error
		}
		c.Response().Header().Set("HX-Redirect", "/listProduct")
		c.Response().WriteHeader(200)
		return c.String(200, "")
	}

	return Render(c, 200, produto.Index(k))
}

func (s Server) EditProduct(c echo.Context) error {
	id := c.Param("ID")
	var p Product
	if err := db.First(&p, id).Error; err != nil {
		return c.String(404, "Produto não encontrado")
	}

	props := &produto.ProductProps{
		ID:                   p.ID,
		Name:                 p.Name,
		Quantity:             p.Quantity,
		Remaining:            p.Remaining,
		Unit:                 p.Unit,
		Date:                 p.Date,
		TotalCost:            p.TotalCost,
		Description:          p.Description,
		PrePulverizationBase: p.PrePulverizationBase,
		Error:                make(map[string]string), // inicializa vazio
	}

	return Render(c, 200, produto.Index(*props))
}

func (s Server) DeleteProduct(c echo.Context) error {
	id := c.Param("ID")
	if err := db.Delete(&Product{}, id).Error; err != nil {
		fmt.Println(err)
		return c.String(500, "Erro ao deletar o produto")
	}

	// Força refresh via HTMX
	c.Response().Header().Set("HX-Refresh", "true")
	return c.NoContent(204)
}
