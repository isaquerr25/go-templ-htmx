package main

import (
	"strconv"

	"github.com/isaquerr25/go-templ-htmx/views/pages/field"
	"github.com/isaquerr25/go-templ-htmx/views/pages/planting"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func GetAllFields() ([]planting.Field, error) {
	var dbFields []Field
	if err := db.Find(&dbFields).Error; err != nil {
		return nil, err
	}

	// Conversão manual para []planting.Field
	var fields []planting.Field
	for _, f := range dbFields {
		fields = append(fields, planting.Field{
			ID:   f.ID,
			Name: f.Name,
		})
	}

	return fields, nil
}

func ShowFieldForm(c echo.Context) error {
	return field.Index(field.FieldProps{}).Render(c.Request().Context(), c.Response().Writer)
}

func CreateField(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var f Field
		if err := c.Bind(&f); err != nil {
			return err
		}
		// validação simples
		errors := map[string]string{}
		if f.Name == "" {
			errors["Name"] = "Nome é obrigatório"
		}
		if f.Hectares <= 0 {
			errors["Hectares"] = "Hectares deve ser maior que zero"
		}
		if len(errors) > 0 {
			return field.Index(field.FieldProps{
				Name:        f.Name,
				Hectares:    f.Hectares,
				Description: f.Description,
				Error:       errors,
			}).Render(c.Request().Context(), c.Response().Writer)
		}
		if err := db.Create(&f).Error; err != nil {
			return err
		}
		return ShowFieldForm(c)
	}
}

func UpdateField(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))
		var f Field
		if err := db.First(&f, id).Error; err != nil {
			return err
		}
		if err := c.Bind(&f); err != nil {
			return err
		}
		if err := db.Save(&f).Error; err != nil {
			return err
		}
		return ShowFieldForm(c)
	}
}

func DeleteField(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))
		if err := db.Delete(&Field{}, id).Error; err != nil {
			return err
		}
		return ShowFieldForm(c)
	}
}

func ListFields(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var fields []Field
		if err := db.Order("created_at desc").Find(&fields).Error; err != nil {
			return err
		}

		var items []field.FieldItem
		for _, f := range fields {
			items = append(items, field.FieldItem{
				ID:          f.ID,
				Name:        f.Name,
				Hectares:    f.Hectares,
				Description: f.Description,
				CreatedAt:   f.CreatedAt,
			})
		}

		return field.List(items).Render(c.Request().Context(), c.Response().Writer)
	}
}
