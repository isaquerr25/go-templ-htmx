package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/isaquerr25/go-templ-htmx/views/pages/sale"
	"github.com/labstack/echo/v4"
)

const (
	MethodCash    SaleMethod = "cash"
	MethodCard    SaleMethod = "card"
	MethodPix     SaleMethod = "pix"
	StatePending  SaleState  = "pending"
	StatePaid     SaleState  = "paid"
	StateCanceled SaleState  = "canceled"
)

// Wrapper para lista de vendas
type SaleListProps struct {
	Sales []SaleListProps
}

func validateSale(
	c echo.Context,
	p *sale.SaleProps,
) (props sale.SaleProps, hasError bool, err error) {
	err = c.Bind(p)
	if err != nil {
		return
	}

	props.Error = map[string]string{}

	p.Unit = strings.TrimSpace(p.Unit)
	p.Method = strings.TrimSpace(p.Method)
	p.State = strings.TrimSpace(p.State)
	p.Notes = strings.TrimSpace(p.Notes)

	if p.ClientID == 0 {
		props.Error["ClientID"] = "ClientID is required"
		hasError = true
	}

	if p.ProductSellID == 0 {
		props.Error["ProductSellID"] = "ProductSellID is required"
		hasError = true
	}

	if p.Quantity <= 0 {
		props.Error["Quantity"] = "Quantity must be greater than zero"
		hasError = true
	}

	if p.Unit == "" {
		props.Error["Unit"] = "Unit is required"
		hasError = true
	}

	if p.Method == "" {
		props.Error["Method"] = "Method is required"
		hasError = true
	}

	if p.State == "" {
		props.Error["State"] = "State is required"
		hasError = true
	}

	// Parse SoldAt string to time.Time, se erro define erro e flag
	soldAt, errTime := time.Parse("2006-01-02", p.SoldAt)
	if errTime != nil {
		props.Error["SoldAt"] = "Invalid date format. Use YYYY-MM-DD"
		hasError = true
	} else {
		// se não erro, sobrescreve valor parseado em p (opcional)
		p.SoldAt = soldAt.Format("2006-01-02")
	}

	props = *p

	return
}

func (s Server) NewSale(c echo.Context) error {
	props := sale.SaleProps{
		Error: make(map[string]string),
	}
	return Render(c, http.StatusOK, sale.Index(props))
}

func (s Server) CreateSale(c echo.Context) error {
	p := &sale.SaleProps{}
	props, hasError, err := validateSale(c, p)
	if err != nil {
		return err
	}

	if !hasError {
		// converter SoldAt para time.Time
		soldAt, _ := time.Parse("2006-01-02", props.SoldAt)

		saleModel := sale.SaleProps{
			ClientID:      props.ClientID,
			ProductSellID: props.ProductSellID,
			SoldAt:        soldAt.String(),
			Quantity:      props.Quantity,
			Unit:          props.Unit,
			TotalPrice:    props.TotalPrice,
			Method:        props.Method,
			State:         props.State,
			Notes:         props.Notes,
		}

		r := db.Create(&saleModel)
		if r.Error != nil {
			return r.Error
		}

		c.Response().Header().Set("HX-Redirect", "/listSale")
		return c.String(http.StatusOK, "")
	}

	return Render(c, http.StatusOK, sale.Index(props))
}

func (s Server) UpdateSale(c echo.Context) error {
	id := c.Param("id")
	p := &sale.SaleProps{}
	props, hasError, err := validateSale(c, p)
	if err != nil {
		return err
	}

	if !hasError {
		var saleModel sale.SaleProps
		if r := db.First(&saleModel, id); r.Error != nil {
			return r.Error
		}

		soldAt, _ := time.Parse("2006-01-02", props.SoldAt)

		saleModel.ClientID = props.ClientID
		saleModel.ProductSellID = props.ProductSellID
		saleModel.SoldAt = soldAt.String()
		saleModel.Quantity = props.Quantity
		saleModel.Unit = props.Unit
		saleModel.TotalPrice = props.TotalPrice
		saleModel.Method = props.Method
		saleModel.State = props.State
		saleModel.Notes = props.Notes

		if r := db.Save(&saleModel); r.Error != nil {
			return r.Error
		}

		c.Response().Header().Set("HX-Redirect", "/listSale")
		return c.String(http.StatusOK, "")
	}

	return Render(c, http.StatusOK, sale.Index(props))
}

func (s Server) DeleteSale(c echo.Context) error {
	id := c.Param("id")

	var saleModel sale.SaleProps
	if r := db.First(&saleModel, id); r.Error != nil {
		return r.Error
	}

	if r := db.Delete(&saleModel); r.Error != nil {
		return r.Error
	}

	c.Response().Header().Set("HX-Redirect", "/listSale")
	return c.String(http.StatusOK, "")
}

func (s Server) ListSale(c echo.Context) error {
	var sales []Sale
	if r := db.Find(&sales); r.Error != nil {
		return r.Error
	}

	propsList := make([]sale.SaleProps, len(sales))
	for i, v := range sales {
		propsList[i] = sale.SaleProps{
			ID:            v.ID,
			ClientID:      *v.ClientID,
			ProductSellID: v.ProductSellID,
			SoldAt:        v.SoldAt.Format("2006-01-02"),
			Quantity:      int(v.Quantity),
			Unit:          v.Unit,
			TotalPrice:    v.TotalPrice,
			Method:        string(v.Method),
			State:         string(v.State),
			Notes:         v.Notes,
			Error:         make(map[string]string),
		}
	}

	props := sale.SaleListProps{
		Sales: propsList,
	}

	return sale.List(props).Render(c.Request().Context(), c.Response().Writer)
}

func (s Server) ShowSale(c echo.Context) error {
	id := c.Param("id")

	var saleModel sale.SaleProps
	if r := db.First(&saleModel, id); r.Error != nil {
		return r.Error
	}

	props := sale.SaleProps{
		ID:            saleModel.ID,
		ClientID:      saleModel.ClientID,
		ProductSellID: saleModel.ProductSellID,
		SoldAt:        saleModel.SoldAt,
		Quantity:      saleModel.Quantity,
		Unit:          saleModel.Unit,
		TotalPrice:    saleModel.TotalPrice,
		Method:        string(saleModel.Method),
		State:         string(saleModel.State),
		Notes:         saleModel.Notes,
	}

	return Render(c, http.StatusOK, sale.Show(props))
}
