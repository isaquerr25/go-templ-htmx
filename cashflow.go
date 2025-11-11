package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/isaquerr25/go-templ-htmx/views/pages/cashflow"
	"github.com/labstack/echo/v4"
)

//

// 📋 LISTAGEM
func ListCashFlows(c echo.Context) error {
	var flows []CashFlow
	if err := db.Find(&flows).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Erro ao buscar fluxos de caixa")
	}

	var items []cashflow.CashFlowProps
	var totalIn, totalOut float64

	for _, f := range flows {
		items = append(items, cashflow.CashFlowProps{
			ID:          f.ID,
			Type:        string(f.Type),
			Category:    string(f.Category),
			Method:      string(f.Method),
			Description: f.Description,
			Amount:      f.Amount,
			OccurredAt:  f.OccurredAt,
		})

		if f.Type == "in" {
			totalIn += f.Amount
		} else if f.Type == "out" {
			totalOut += f.Amount
		}
	}

	balance := totalIn - totalOut

	props := cashflow.CashFlowListProps{
		Items:    items,
		TotalIn:  totalIn,
		TotalOut: totalOut,
		Balance:  balance,
	}

	return cashflow.List(props).Render(c.Request().Context(), c.Response())
}

// 📄 FORM DE CRIAÇÃO
func ShowCreateCashFlow(c echo.Context) error {
	p := cashflow.CashFlowProps{
		Error:      map[string]string{},
		OccurredAt: time.Now(), // 👉 define a data atual por padrão
	}

	return cashflow.Index(p).Render(c.Request().Context(), c.Response())
}

// 🆕 CRIAR
func CreateCashFlow(c echo.Context) error {
	req := c.Request()
	if err := req.ParseForm(); err != nil {
		return c.String(http.StatusBadRequest, "Erro ao processar formulário")
	}
	form := req.Form

	input := cashflow.CashFlowProps{
		Error: map[string]string{},
	}

	input.Type = form.Get("type")
	input.Category = form.Get("category")
	input.Method = form.Get("method")
	input.Description = form.Get("description")
	input.Notes = form.Get("notes")

	amount, err := strconv.ParseFloat(form.Get("amount"), 64)
	if err != nil || amount <= 0 {
		input.Error["Amount"] = "Valor inválido"
	}
	input.Amount = amount

	occurredAt, err := time.Parse("2006-01-02T15:04", form.Get("occurredAt"))
	if err != nil {
		input.Error["OccurredAt"] = "Data inválida"
	} else {
		input.OccurredAt = occurredAt
	}

	if input.Type == "" {
		input.Error["Type"] = "Informe o tipo (entrada/saída)"
	}

	if len(input.Error) > 0 {
		return cashflow.Index(input).Render(c.Request().Context(), c.Response())
	}

	flow := CashFlow{
		Type:        FlowType(input.Type),
		Category:    FlowCategory(input.Category),
		Amount:      input.Amount,
		Method:      FlowMethod(input.Method),
		OccurredAt:  input.OccurredAt,
		Description: input.Description,
		Notes:       input.Notes,
	}

	if err := db.Create(&flow).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Erro ao salvar movimentação")
	}

	c.Response().Header().Set("HX-Redirect", "/cashflow")
	return c.String(http.StatusOK, "")
}

// ✏️ MOSTRAR PARA EDIÇÃO
func ShowCashFlow(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var f CashFlow

	if err := db.First(&f, id).Error; err != nil {
		return c.String(http.StatusNotFound, "Movimentação não encontrada")
	}

	p := cashflow.CashFlowProps{
		ID:          f.ID,
		Type:        string(f.Type),
		Category:    string(f.Category),
		Amount:      f.Amount,
		Method:      string(f.Method),
		OccurredAt:  f.OccurredAt,
		Description: f.Description,
		Notes:       f.Notes,
		Error:       map[string]string{},
	}

	return cashflow.Index(p).Render(c.Request().Context(), c.Response())
}

// 🔄 ATUALIZAR
func UpdateCashFlow(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var f CashFlow

	if err := db.First(&f, id).Error; err != nil {
		return c.String(http.StatusNotFound, "Movimentação não encontrada")
	}

	req := c.Request()
	if err := req.ParseForm(); err != nil {
		return c.String(http.StatusBadRequest, "Erro ao processar formulário")
	}
	form := req.Form

	f.Type = FlowType(form.Get("type"))
	f.Category = FlowCategory(form.Get("category"))
	f.Method = FlowMethod(form.Get("method"))
	f.Description = form.Get("description")
	f.Notes = form.Get("notes")

	if val, err := strconv.ParseFloat(form.Get("amount"), 64); err == nil {
		f.Amount = val
	}

	if t, err := time.Parse("2006-01-02T15:04", form.Get("occurredAt")); err == nil {
		f.OccurredAt = t
	}

	if err := db.Save(&f).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Erro ao atualizar movimentação")
	}

	c.Response().Header().Set("HX-Redirect", "/cashflow")
	return c.String(http.StatusOK, "")
}

// ❌ DELETAR
func DeleteCashFlow(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "ID inválido")
	}

	if err := db.Delete(&CashFlow{}, id).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Erro ao deletar movimentação")
	}

	c.Response().Header().Set("HX-Redirect", "")
	return c.String(http.StatusOK, "")
}
