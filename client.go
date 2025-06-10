package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/isaquerr25/go-templ-htmx/views/pages/client"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type ClientProps struct {
	ID      uint
	Name    string
	Email   string
	Phone   string
	Company string
	Address string
	Notes   string
	Error   map[string]string
}

func GetAllClientsProps(db *gorm.DB) ([]ClientProps, error) {
	var clients []Client
	if err := db.Find(&clients).Error; err != nil {
		return nil, err
	}

	// Mapear para ClientProps
	var clientPropsList []ClientProps
	for _, c := range clients {
		clientPropsList = append(clientPropsList, ClientProps{
			ID:      c.ID,
			Name:    c.Name,
			Email:   c.Email,
			Phone:   c.Phone,
			Company: c.Company,
			Address: c.Address,
			Notes:   c.Notes,
			Error:   map[string]string{}, // vazio por padrão
		})
	}

	return clientPropsList, nil
}

func validateClient(
	c echo.Context,
	p *client.ClientProps,
) (props client.ClientProps, hasError bool, err error) {
	err = c.Bind(p)
	if err != nil {
		return
	}

	props.Error = map[string]string{}

	p.Name = strings.TrimSpace(p.Name)
	p.Email = strings.TrimSpace(p.Email)
	p.Phone = strings.TrimSpace(p.Phone)
	p.Company = strings.TrimSpace(p.Company)
	p.Address = strings.TrimSpace(p.Address)
	p.Notes = strings.TrimSpace(p.Notes)

	if p.Name == "" {
		props.Error["Name"] = "Name is required"
		hasError = true
	}
	if p.Email == "" {
		props.Error["Email"] = "Email is required"
		hasError = true
	}
	if p.Phone == "" {
		props.Error["Phone"] = "Phone is required"
		hasError = true
	}

	props.ID = p.ID
	props.Name = p.Name
	props.Email = p.Email
	props.Phone = p.Phone
	props.Company = p.Company
	props.Address = p.Address
	props.Notes = p.Notes

	return
}

func (s Server) ListClient(c echo.Context) error {
	var clients []Client
	if err := db.Find(&clients).Error; err != nil {
		return err
	}

	// Map Client to ClientProps
	propsList := make([]client.ClientProps, 0, len(clients))
	for _, cl := range clients {
		propsList = append(propsList, client.ClientProps{
			ID:      cl.ID,
			Name:    cl.Name,
			Email:   cl.Email,
			Phone:   cl.Phone,
			Company: cl.Company,
			Address: cl.Address,
			Notes:   cl.Notes,
			Error:   nil,
		})
	}

	return Render(c, 200, client.List(client.ClientListProps{Clients: propsList}))
}

func (s Server) ShowClient(c echo.Context) error {
	id := c.Param("id")
	var cl Client
	if err := db.First(&cl, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.NoContent(http.StatusNotFound)
		}
		return err
	}

	props := client.ClientProps{
		ID:      cl.ID,
		Name:    cl.Name,
		Email:   cl.Email,
		Phone:   cl.Phone,
		Company: cl.Company,
		Address: cl.Address,
		Notes:   cl.Notes,
		Error:   nil,
	}

	return Render(c, http.StatusOK, client.Index(props))
}

func (s Server) CreateClient(c echo.Context) error {
	clProps := &client.ClientProps{}
	props, hasError, err := validateClient(c, clProps)
	if err != nil {
		return err
	}

	if !hasError {
		cl := Client{
			Name:    clProps.Name,
			Email:   clProps.Email,
			Phone:   clProps.Phone,
			Company: clProps.Company,
			Address: clProps.Address,
			Notes:   clProps.Notes,
		}
		if err := db.Create(&cl).Error; err != nil {
			fmt.Println(err)
			return err
		}
		c.Response().Header().Set("HX-Redirect", "/listCustomer")
		return c.NoContent(http.StatusOK)
	}

	return Render(c, http.StatusOK, client.Index(props))
}

func (s Server) UpdateClient(c echo.Context) error {
	id := c.Param("id")
	var cl Client
	if err := db.First(&cl, id).Error; err != nil {
		return err
	}

	// Cria props com os dados atuais do client
	clProps := &client.ClientProps{
		ID:      cl.ID,
		Name:    cl.Name,
		Email:   cl.Email,
		Phone:   cl.Phone,
		Company: cl.Company,
		Address: cl.Address,
		Notes:   cl.Notes,
	}

	props, hasError, err := validateClient(c, clProps)
	if err != nil {
		return err
	}

	if !hasError {
		cl.Name = clProps.Name
		cl.Email = clProps.Email
		cl.Phone = clProps.Phone
		cl.Company = clProps.Company
		cl.Address = clProps.Address
		cl.Notes = clProps.Notes

		if err := db.Save(&cl).Error; err != nil {
			return err
		}
		c.Response().Header().Set("HX-Redirect", "/listCustomer")
		return c.NoContent(http.StatusOK)
	}

	return Render(c, http.StatusOK, client.Index(props))
}

func (s Server) DeleteClient(c echo.Context) error {
	id := c.Param("id")
	if err := db.Delete(&Client{}, id).Error; err != nil {
		return err
	}
	c.Response().Header().Set("HX-Redirect", "/listCustomer")
	return c.NoContent(http.StatusOK)
}
