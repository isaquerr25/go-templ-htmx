package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/isaquerr25/go-templ-htmx/views/pages/categoria"
	"github.com/labstack/echo/v4"
)

func HandleListCategorias(c echo.Context) error {
	nameFilter := strings.TrimSpace(c.QueryParam("name"))

	var categorias []Categoria
	query := db.Model(&Categoria{})
	if nameFilter != "" {
		query = query.Where("LOWER(name) LIKE LOWER(?)", "%"+nameFilter+"%")
	}
	query.Order("name ASC").Find(&categorias)

	var items []categoria.CategoriaItem
	for _, ct := range categorias {
		items = append(items, categoria.CategoriaItem{
			ID:   ct.ID,
			Name: ct.Name,
		})
	}

	if c.Request().Header.Get("HX-Request") == "true" {
		return Render(c, http.StatusOK, categoria.ListContent(items, nameFilter))
	}
	return Render(c, http.StatusOK, categoria.List(items, nameFilter))
}

func HandleShowCategoriaForm(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return Render(c, http.StatusOK, categoria.Form(categoria.CategoriaProps{
			Error: map[string]string{},
		}))
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "ID inválido")
	}

	var ct Categoria
	if err := db.First(&ct, id).Error; err != nil {
		return c.String(http.StatusNotFound, "Categoria não encontrada")
	}

	return Render(c, http.StatusOK, categoria.Form(categoria.CategoriaProps{
		ID:    ct.ID,
		Name:  ct.Name,
		Error: map[string]string{},
	}))
}

func HandleCreateCategoria(c echo.Context) error {
	name := strings.TrimSpace(c.FormValue("name"))
	errorsMap := map[string]string{}

	if name == "" {
		errorsMap["Name"] = "Nome da categoria é obrigatório"
	}

	if len(errorsMap) > 0 {
		return Render(c, http.StatusOK, categoria.Form(categoria.CategoriaProps{
			Name:  name,
			Error: errorsMap,
		}))
	}

	var existing Categoria
	if err := db.Where("LOWER(name) = LOWER(?)", name).First(&existing).Error; err == nil {
		errorsMap["Name"] = "Já existe uma categoria com este nome"
		return Render(c, http.StatusOK, categoria.Form(categoria.CategoriaProps{
			Name:  name,
			Error: errorsMap,
		}))
	}

	ct := Categoria{Name: name}
	if err := db.Create(&ct).Error; err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	c.Response().Header().Set("HX-Redirect", "/categorias")
	return c.NoContent(http.StatusOK)
}

func HandleUpdateCategoria(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "ID inválido")
	}

	var ct Categoria
	if err := db.First(&ct, id).Error; err != nil {
		return c.String(http.StatusNotFound, "Categoria não encontrada")
	}

	name := strings.TrimSpace(c.FormValue("name"))
	errorsMap := map[string]string{}

	if name == "" {
		errorsMap["Name"] = "Nome da categoria é obrigatório"
	}

	if len(errorsMap) > 0 {
		return Render(c, http.StatusOK, categoria.Form(categoria.CategoriaProps{
			ID:    ct.ID,
			Name:  name,
			Error: errorsMap,
		}))
	}

	var existing Categoria
	if err := db.Where("LOWER(name) = LOWER(?) AND id <> ?", name, id).First(&existing).Error; err == nil {
		errorsMap["Name"] = "Já existe uma categoria com este nome"
		return Render(c, http.StatusOK, categoria.Form(categoria.CategoriaProps{
			ID:    ct.ID,
			Name:  name,
			Error: errorsMap,
		}))
	}

	ct.Name = name
	if err := db.Save(&ct).Error; err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	c.Response().Header().Set("HX-Redirect", "/categorias")
	return c.NoContent(http.StatusOK)
}

func HandleDeleteCategoria(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "ID inválido")
	}

	result := db.Delete(&Categoria{}, id)
	if result.Error != nil {
		return c.String(http.StatusInternalServerError, "Erro ao deletar")
	}

	c.Response().Header().Set("HX-Refresh", "true")
	return c.NoContent(http.StatusOK)
}
