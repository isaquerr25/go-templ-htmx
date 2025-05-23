package main

import (
	"embed"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/isaquerr25/go-templ-htmx/views/pages/client"
	"github.com/isaquerr25/go-templ-htmx/views/pages/fertilization"
	"github.com/isaquerr25/go-templ-htmx/views/pages/field"
	"github.com/isaquerr25/go-templ-htmx/views/pages/harvest"
	"github.com/isaquerr25/go-templ-htmx/views/pages/home"
	"github.com/isaquerr25/go-templ-htmx/views/pages/planting"
	"github.com/isaquerr25/go-templ-htmx/views/pages/produto"
	"github.com/isaquerr25/go-templ-htmx/views/pages/sale"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//go:embed static/*
var assets embed.FS

var db *gorm.DB

func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	if err := t.Render(ctx.Request().Context(), buf); err != nil {
		return err
	}

	return ctx.HTML(statusCode, buf.String())
}

func validateProduct(c echo.Context, p *Product) (
	k produto.ProductProps,
	hasError bool,
	err error,
) {
	var values map[string]string

	// Bind request parameters to values map
	err = c.Bind(&values)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Trim spaces and assign to the product fields
	p.Name = strings.TrimSpace(values["name"])
	p.Description = strings.TrimSpace(values["description"])
	p.Unit = strings.TrimSpace(values["unit"])

	// Validate Quantity
	quantity, err := strconv.ParseFloat(values["quantity"], 64)
	if err != nil {
		k.Error["Quantity"] = "Only numbers are allowed"
		hasError = true
	} else if quantity <= 0 {
		k.Error["Quantity"] = "Cannot be less than or equal to zero"
		hasError = true
	}
	p.Quantity = quantity

	// Validate TotalCost
	totalCost, err := strconv.ParseFloat(values["totalCost"], 64)
	if err != nil {
		k.Error["TotalCost"] = "Only numbers are allowed"
		hasError = true
	} else if totalCost < 1 {
		k.Error["TotalCost"] = "Cannot be less than 1"
		hasError = true
	}
	p.TotalCost = totalCost

	// Validate Date
	date, err := time.Parse("2006-01-02", values["date"])
	if err != nil {
		k.Error["Date"] = "Invalid date format"
		hasError = true
	}
	p.Date = date

	// Validate Name
	if strings.TrimSpace(values["name"]) == "" {
		k.Error["Name"] = "Name is required"
		hasError = true
	}

	// Populate ProductProps with validated data
	k.Name = p.Name
	k.Quantity = p.Quantity
	k.Unit = p.Unit
	k.Date = p.Date
	k.TotalCost = p.TotalCost
	k.Description = p.Description

	return
}

func main() {
	e := echo.New()

	s := Server{}

	var err error
	db, err = gorm.Open(sqlite.Open("base.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&Product{})
	db.AutoMigrate(&Field{})
	db.AutoMigrate(&Planting{})
	db.AutoMigrate(&Fertilization{})
	db.AutoMigrate(&Pulverization{})
	db.AutoMigrate(&Irrigation{})
	db.AutoMigrate(&Client{})
	db.AutoMigrate(&ProductSell{})
	db.AutoMigrate(&Sale{})
	db.AutoMigrate(&Harvest{})

	e.Static("/static", "static")

	e.POST("/updateProduct/:ID", s.UpdateProduct)

	e.POST("/createProduct", func(c echo.Context) error {
		p := &Product{}
		k, hasError, err := validateProduct(c, p)
		if err != nil {
			return err
		}

		if !hasError {
			r := db.Create(&p)
			if r.Error != nil {
				fmt.Println(r.Error)

				return err
			}
			c.Response().Header().Set("HX-Redirect", "/listProduct")
			c.Response().WriteHeader(200)
			return c.String(200, "")
		}

		return Render(c, 200, produto.Index(k))
	})

	e.GET("/listProduct/new",
		func(c echo.Context) error {
			return Render(c, 200, produto.Index(produto.ProductProps{
				Quantity:    1,
				Date:        time.Now(),
				Error:       map[string]string{},
				Name:        "",
				Remaining:   0,
				Unit:        "",
				TotalCost:   150,
				Description: "",
			}))
		},
	)

	e.GET("/listProduct",

		func(c echo.Context) error {
			var pp []Product

			r := db.Find(&pp)
			if r.Error != nil {
				fmt.Println(r.Error)
				return err
			}

			props := make([]produto.ProductProps, len(pp))

			for i, p := range pp {
				props[i] = produto.ProductProps{
					ID:          p.ID,
					Name:        p.Name,
					Quantity:    p.Quantity,
					Remaining:   p.Remaining,
					Unit:        p.Unit,
					Date:        p.Date,
					TotalCost:   p.TotalCost,
					Description: p.Description,
					Error:       map[string]string{},
				}
			}
			return Render(c, 200, produto.List(props))
		},
	)

	e.GET("/listProduct/:ID",
		func(c echo.Context) error {
			var p Product

			r := db.First(&p, c.Param("ID"))
			if r.Error != nil {
				fmt.Println(r.Error)
				return err
			}

			props := produto.ProductProps{
				ID:          p.ID,
				Name:        p.Name,
				Quantity:    p.Quantity,
				Remaining:   p.Remaining,
				Unit:        p.Unit,
				Date:        p.Date,
				TotalCost:   p.TotalCost,
				Description: p.Description,
				Error:       map[string]string{},
			}

			return Render(c, 200, produto.Index(props))
		},
	)

	e.GET("/plantings", ListPlantings(db))
	e.GET("/plantings/new", ShowPlantingForm(db))
	e.GET("/plantings/edit/:id", ShowPlantingForm(db))
	e.POST("/plantings/create", CreatePlanting(db))
	e.POST("/plantings/update/:id", UpdatePlanting(db))
	e.DELETE("/plantings/delete/:id", DeletePlanting(db))

	e.GET("/plantings/list", ListPlantings(db))

	e.GET("/plantings/edit/:id", func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))
		var p Planting
		if err := db.First(&p, id).Error; err != nil {
			return err
		}
		return planting.Index(planting.PlantingProps{
			ID:          p.ID,
			FieldID:     p.FieldID,
			CropName:    p.CropName,
			StartedAt:   p.StartedAt.Format("2006-01-02"),
			EndedAt:     nullableDate(p.EndedAt),
			IsCompleted: p.IsCompleted,
			AreaUsed:    p.AreaUsed,
		}).Render(c.Request().Context(), c.Response().Writer)
	})

	// Rotas existentes
	e.GET("/listCustomer", s.ListClient)
	e.GET("/showClient/:id", s.ShowClient)
	e.POST("/createClient", s.CreateClient)
	e.POST("/updateClient/:id", s.UpdateClient)
	e.POST("/deleteClient/:id", s.DeleteClient)
	e.GET("/listCustomer/new", func(c echo.Context) error {
		return Render(c, 200, client.Index(client.ClientProps{
			Error: map[string]string{},
		}))
	})

	e.GET("/listSale", s.ListSale)

	e.GET("/createSale", func(c echo.Context) error {
		return Render(c, http.StatusOK, sale.Index(sale.SaleProps{}))
	})
	e.POST("/createSale", s.CreateSale)
	e.GET("/updateSale/:id", func(c echo.Context) error {
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
		return Render(c, http.StatusOK, sale.Index(props))
	})
	e.POST("/updateSale/:id", s.UpdateSale)
	e.POST("/deleteSale/:id", s.DeleteSale)
	e.GET("/showSale/:id", s.ShowSale)

	e.GET("/fields", ShowFieldForm)
	e.POST("/fields/create", CreateField(db))
	e.POST("/fields/update/:id", UpdateField(db))
	e.DELETE("/fields/delete/:id", DeleteField(db))

	e.GET("/fields/list", ListFields(db))
	e.GET("/fields/edit/:id", func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))
		var f Field
		if err := db.First(&f, id).Error; err != nil {
			return err
		}
		return field.Index(field.FieldProps{
			ID:          f.ID,
			Name:        f.Name,
			Hectares:    f.Hectares,
			Description: f.Description,
		}).Render(c.Request().Context(), c.Response().Writer)
	})

	e.GET("/pulverizations", ListPulverizations(db))

	e.GET("/pulverization", ShowPulverizationForm(db))
	e.GET("/pulverization/:id", ShowPulverizationForm(db))
	e.POST("/pulverization", CreatePulverization(db))
	e.POST("/pulverization/:id", UpdatePulverization(db))
	e.DELETE("/pulverization/:id", DeletePulverization(db))

	e.GET("/irrigation", IrrigationIndex)
	e.POST("/irrigation/create", IrrigationCreate)
	e.GET("/irrigation/list", IrrigationList)
	e.GET("/irrigation/edit/:id", IrrigationEdit)

	e.GET("/harvest", ListHarvest)
	e.GET("/harvest/:id", ShowHarvest)
	e.GET("/harvest/create", func(c echo.Context) error {
		return harvest.Index(harvest.HarvestProps{}).Render(c.Request().Context(), c.Response())
	})
	e.POST("/harvest/create", CreateHarvest)
	e.POST("/harvest/update/:id", UpdateHarvest)

	// Fertilization routes
	e.GET("/fertilization", ListFertilization)
	e.GET("/fertilization/create", func(c echo.Context) error {
		return fertilization.Index(fertilization.FertilizationProps{}).
			Render(c.Request().Context(), c.Response())
	})
	e.GET("/fertilization/:id", ShowFertilization)
	e.POST("/fertilization/create", CreateFertilization)
	e.POST("/fertilization/update/:id", UpdateFertilization)

	// Rotas de vendas (sale)
	e.GET("/sales", s.ListSale)          // Lista todas as vendas
	e.GET("/sales/:id", s.ShowSale)      // Mostra detalhes de uma venda
	e.POST("/sales", s.CreateSale)       // Cria uma nova venda
	e.POST("/sales/:id", s.UpdateSale)   // Atualiza uma venda existente
	e.DELETE("/sales/:id", s.DeleteSale) // Deleta uma venda

	e.GET("/productsell", ListProductSell)
	e.GET("/productsell/create", CreateViewProductSell)
	e.POST("/productsell/create", CreateProductSell)
	e.GET("/productsell/:id", EditViewProductSell)
	e.POST("/productsell/update/:id", UpdateProductSell)
	e.GET("/",

		func(c echo.Context) error {
			return Render(c, 200, home.Hello("asds"))
		},
	)
	e.Logger.Fatal(e.Start(":1323"))
}

func nullableDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}
