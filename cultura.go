package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/isaquerr25/go-templ-htmx/views/pages/cultura"
	"github.com/isaquerr25/go-templ-htmx/views/pages/produto"
	"github.com/labstack/echo/v4"
)

// ========================
// LIST CULTURAS
// ========================
func HandleListCulturas(c echo.Context) error {
	nameFilter := strings.TrimSpace(c.QueryParam("name"))

	var culturas []Cultura
	query := db.Model(&Cultura{})
	if nameFilter != "" {
		query = query.Where("LOWER(name) LIKE LOWER(?)", "%"+nameFilter+"%")
	}
	query.Order("name ASC").Find(&culturas)

	var items []cultura.CulturaItem
	for _, ct := range culturas {
		items = append(items, cultura.CulturaItem{
			ID:                ct.ID,
			Name:              ct.Name,
			GerminacaoInicio:  ct.GerminacaoInicio,
			FloracaoInicio:    ct.FloracaoInicio,
			ColheitaInicio:    ct.ColheitaInicio,
			MorteInicio:       ct.MorteInicio,
		})
	}

	if c.Request().Header.Get("HX-Request") == "true" {
		return Render(c, http.StatusOK, cultura.ListContent(items, nameFilter))
	}
	return Render(c, http.StatusOK, cultura.List(items, nameFilter))
}

// ========================
// SHOW CULTURA FORM (NEW/EDIT)
// ========================
func HandleShowCulturaForm(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return Render(c, http.StatusOK, cultura.Form(cultura.CulturaProps{
			Error: map[string]string{},
		}))
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "ID inválido")
	}

	var ct Cultura
	if err := db.First(&ct, id).Error; err != nil {
		return c.String(http.StatusNotFound, "Cultura não encontrada")
	}

	return Render(c, http.StatusOK, cultura.Form(cultura.CulturaProps{
		ID:                ct.ID,
		Name:              ct.Name,
		GerminacaoInicio:  ct.GerminacaoInicio,
		FloracaoInicio:    ct.FloracaoInicio,
		ColheitaInicio:    ct.ColheitaInicio,
		MorteInicio:       ct.MorteInicio,
		Error:             map[string]string{},
	}))
}

// ========================
// CREATE CULTURA
// ========================
func HandleCreateCultura(c echo.Context) error {
	name := strings.TrimSpace(c.FormValue("name"))
	errorsMap := map[string]string{}

	if name == "" {
		errorsMap["Name"] = "Nome da cultura é obrigatório"
	}

	if len(errorsMap) > 0 {
		return Render(c, http.StatusOK, cultura.Form(cultura.CulturaProps{
			Name:  name,
			Error: errorsMap,
		}))
	}

	var existing Cultura
	if err := db.Where("LOWER(name) = LOWER(?)", name).First(&existing).Error; err == nil {
		errorsMap["Name"] = "Já existe uma cultura com este nome"
		return Render(c, http.StatusOK, cultura.Form(cultura.CulturaProps{
			Name:  name,
			Error: errorsMap,
		}))
	}

	germinacao, _ := strconv.Atoi(c.FormValue("germinacaoInicio"))
	floracao, _ := strconv.Atoi(c.FormValue("floracaoInicio"))
	colheita, _ := strconv.Atoi(c.FormValue("colheitaInicio"))
	morte, _ := strconv.Atoi(c.FormValue("morteInicio"))

	ct := Cultura{
		Name:             name,
		GerminacaoInicio: germinacao,
		FloracaoInicio:   floracao,
		ColheitaInicio:   colheita,
		MorteInicio:      morte,
	}
	if err := db.Create(&ct).Error; err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	c.Response().Header().Set("HX-Redirect", "/culturas")
	return c.NoContent(http.StatusOK)
}

// ========================
// UPDATE CULTURA
// ========================
func HandleUpdateCultura(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "ID inválido")
	}

	var ct Cultura
	if err := db.First(&ct, id).Error; err != nil {
		return c.String(http.StatusNotFound, "Cultura não encontrada")
	}

	name := strings.TrimSpace(c.FormValue("name"))
	errorsMap := map[string]string{}

	if name == "" {
		errorsMap["Name"] = "Nome da cultura é obrigatório"
	}

	if len(errorsMap) > 0 {
		return Render(c, http.StatusOK, cultura.Form(cultura.CulturaProps{
			ID:    ct.ID,
			Name:  name,
			Error: errorsMap,
		}))
	}

	var existing Cultura
	if err := db.Where("LOWER(name) = LOWER(?) AND id <> ?", name, id).First(&existing).Error; err == nil {
		errorsMap["Name"] = "Já existe uma cultura com este nome"
		return Render(c, http.StatusOK, cultura.Form(cultura.CulturaProps{
			ID:    ct.ID,
			Name:  name,
			Error: errorsMap,
		}))
	}

	ct.Name = name
	ct.GerminacaoInicio, _ = strconv.Atoi(c.FormValue("germinacaoInicio"))
	ct.FloracaoInicio, _ = strconv.Atoi(c.FormValue("floracaoInicio"))
	ct.ColheitaInicio, _ = strconv.Atoi(c.FormValue("colheitaInicio"))
	ct.MorteInicio, _ = strconv.Atoi(c.FormValue("morteInicio"))
	if err := db.Save(&ct).Error; err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	c.Response().Header().Set("HX-Redirect", "/culturas")
	return c.NoContent(http.StatusOK)
}

// ========================
// DELETE CULTURA
// ========================
func HandleDeleteCultura(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "ID inválido")
	}

	result := db.Delete(&Cultura{}, id)
	if result.Error != nil {
		return c.String(http.StatusInternalServerError, "Erro ao deletar")
	}

	c.Response().Header().Set("HX-Refresh", "true")
	return c.NoContent(http.StatusOK)
}

// ========================
// PRODUCT CULTURA MODAL
// ========================
func HandleEditCulturesModal(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "ID inválido")
	}

	rows := getAllCulturasForEdit(uint(productID))
	return Render(c, http.StatusOK, produto.EditCulturesModal(rows, uint(productID)))
}

// ========================
// SAVE CULTURES (bulk upsert)
// ========================
func HandleSaveCultures(c echo.Context) error {
	productID, err := strconv.Atoi(c.FormValue("product_id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "ID do produto inválido")
	}

	var product Product
	if err := db.First(&product, productID).Error; err != nil {
		return c.String(http.StatusNotFound, "Produto não encontrado")
	}

	form := c.Request().PostForm
	i := 0
	for {
		culturaIDStr := form.Get(fmt.Sprintf("cultures[%d].cultura_id", i))
		if culturaIDStr == "" {
			break
		}

		culturaID, err := strconv.Atoi(culturaIDStr)
		if err != nil {
			i++
			continue
		}

		proportionStr := form.Get(fmt.Sprintf("cultures[%d].proportion", i))
		proportion, parseErr := strconv.ParseFloat(proportionStr, 64)

		if proportionStr == "" || parseErr != nil || proportion <= 0 {
			db.Unscoped().Where("product_id = ? AND cultura_id = ?", productID, culturaID).Delete(&ProductCultura{})
		} else {
			var existing ProductCultura
			err := db.Unscoped().Where("product_id = ? AND cultura_id = ?", productID, culturaID).First(&existing).Error
			if err != nil {
				db.Create(&ProductCultura{
					ProductID:  uint(productID),
					CulturaID:  uint(culturaID),
					Proportion: proportion,
				})
			} else {
				db.Model(&existing).Update("proportion", proportion)
			}
		}

		i++
	}

	c.Response().Header().Set("HX-Redirect", fmt.Sprintf("/editProduct/%d", productID))
	return c.NoContent(http.StatusOK)
}
