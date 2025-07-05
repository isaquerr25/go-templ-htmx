package main

import (
	"fmt"

	"github.com/isaquerr25/go-templ-htmx/views/pages/typeproduct"
	"github.com/labstack/echo/v4"
)

func GetAllTypeProductProps() ([]typeproduct.TypeProductProps, error) {
	var models []TypeProduct
	if err := db.Find(&models).Error; err != nil {
		return nil, err
	}

	var props []typeproduct.TypeProductProps
	for _, m := range models {
		props = append(props, typeproduct.TypeProductProps{
			ID:       m.ID,
			Name:     m.Name,
			Describe: m.Describe,
			Quantity: m.Quantity,
			Error:    map[string]string{},
		})
	}
	return props, nil
}

func (s Server) EditTypeProduct(c echo.Context) error {
	id := c.Param("ID")
	var m TypeProduct
	if err := db.First(&m, id).Error; err != nil {
		return c.String(404, "typeProduct não encontrado")
	}

	props := typeproduct.TypeProductProps{
		ID:       m.ID,
		Name:     m.Name,
		Describe: m.Describe,
		Error:    map[string]string{},
	}

	return Render(c, 200, typeproduct.Index(props))
}

func (s Server) DeleteTypeProduct(c echo.Context) error {
	id := c.Param("ID")
	if err := db.Delete(&TypeProduct{}, id).Error; err != nil {
		fmt.Println(err)
		return c.String(500, "Erro ao deletar typeProduct")
	}

	c.Response().Header().Set("HX-Refresh", "true")
	return c.NoContent(204)
}

func (s Server) UpdateTypeProduct(c echo.Context) error {
	var m TypeProduct
	if err := db.First(&m, c.Param("ID")).Error; err != nil {
		fmt.Println(err)
		return c.String(404, "typeProduct não encontrado")
	}

	props, hasError, err := validateTypeProduct(c, &m)
	if err != nil {
		return err
	}

	if !hasError {
		if err := db.Save(&m).Error; err != nil {
			fmt.Println(err)
			return c.String(500, "Erro ao atualizar")
		}

		c.Response().Header().Set("HX-Redirect", "/listTypeProduct")
		return c.NoContent(200)
	}

	return Render(c, 200, typeproduct.Index(props))
}

func (s Server) CreateTypeProduct(c echo.Context) error {
	var m TypeProduct
	props, hasError, err := validateTypeProduct(c, &m)
	if err != nil {
		return err
	}

	if !hasError {
		if err := db.Create(&m).Error; err != nil {
			fmt.Println(err)
			return c.String(500, "Erro ao criar")
		}
		c.Response().Header().Set("HX-Redirect", "/listTypeProduct")
		return c.NoContent(200)
	}

	return Render(c, 200, typeproduct.Index(props))
}

func validateTypeProduct(
	c echo.Context,
	m *TypeProduct,
) (typeproduct.TypeProductProps, bool, error) {
	var props typeproduct.TypeProductProps
	errors := make(map[string]string)

	if err := c.Bind(m); err != nil {
		return props, true, err
	}

	props = typeproduct.TypeProductProps{
		ID:       m.ID,
		Name:     m.Name,
		Describe: m.Describe,
		Error:    errors,
	}

	if m.Name == "" {
		errors["Name"] = "Nome é obrigatório"
	}
	if m.Describe == "" {
		errors["Describe"] = "Descrição é obrigatória"
	}

	hasError := len(errors) > 0
	props.Error = errors

	return props, hasError, nil
}

func (s Server) ListTypeProduct(c echo.Context) error {
	items, err := GetAllTypeProductProps()
	if err != nil {
		return c.String(500, fmt.Sprintln(err))
	}

	return Render(c, 200, typeproduct.List(typeproduct.TypeProductListProps{
		Items: items,
	}))
}
