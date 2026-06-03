package main

import (
	"embed"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/isaquerr25/go-templ-htmx/views/pages/client"
	"github.com/isaquerr25/go-templ-htmx/views/pages/fertilization"
	"github.com/isaquerr25/go-templ-htmx/views/pages/field"
	"github.com/isaquerr25/go-templ-htmx/views/pages/harvest"
	"github.com/isaquerr25/go-templ-htmx/views/pages/produto"
	"github.com/isaquerr25/go-templ-htmx/views/pages/pulverization"
	"github.com/isaquerr25/go-templ-htmx/views/pages/sale"
	"github.com/isaquerr25/go-templ-htmx/views/pages/typeproduct"
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

func main() {
	e := echo.New()

	s := Server{}

	var err error
	db, err = gorm.Open(sqlite.Open("base.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&CashFlow{})
	db.AutoMigrate(&Service{})
	db.AutoMigrate(&ApplyFertilization{})

	db.AutoMigrate(&AppliedProduct{})
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
	db.AutoMigrate(&TypeProduct{})
	db.AutoMigrate(&Vaccination{})
	db.AutoMigrate(&ProductLot{})
	db.AutoMigrate(&Cultura{})
	db.AutoMigrate(&ProductCultura{})
	db.AutoMigrate(&Categoria{})
	db.AutoMigrate(&ProductCategoria{})
	seedCategorias()

	e.Static("/static", "static")

	// ========================
	// PRODUCT ROUTES
	// ========================
	e.GET("/listProduct", HandleListProduct)
	e.GET("/listProduct/:ID", HandleShowProduct)
	e.GET("/newProduct", HandleNewProduct)
	e.POST("/createProduct", HandleCreateProduct)
	e.POST("/updateProduct/:ID", HandleUpdateProduct)
	e.GET("/editProduct/:ID", HandleEditProduct)
	e.DELETE("/deleteProduct/:ID", HandleDeleteProduct)
	e.DELETE("/product/remove/:id", HandleRemoveProductVisual)

	// LOT ROUTES
	e.GET("/product/:id/add-lot", HandleAddLot)
	e.POST("/product/:id/create-lot", HandleCreateLot)
	e.GET("/editLot/:id", HandleEditLot)
	e.PUT("/updateLot/:id", HandleUpdateLot)
	e.DELETE("/deleteLot/:id", HandleDeleteLot)

	// PRODUCT-CULTURA ROUTES
	e.GET("/product/:id/culturas/edit", HandleEditCulturesModal)
	e.POST("/product/:id/culturas/save", HandleSaveCultures)

	// ========================
	// PULVERIZATION SHARED
	// ========================
	e.GET("/product/showNewInstace", func(c echo.Context) error {
		jj, _ := strconv.Atoi(c.QueryParam("index"))

		a, _ := GetAllProductsForUserProps()
		return Render(
			c,
			200,
			pulverization.ItemsProdut(jj, pulverization.ProductInput{}, pulverization.UseProps{
				Prod: a,
			}),
		)
	})

	e.GET("/plan/showNewInstacePlants", func(c echo.Context) error {
		jj, _ := strconv.Atoi(c.QueryParam("index"))

		a, _ := GetAllPlantings()
		return Render(
			c,
			200,
			pulverization.ItemsPlants(jj, pulverization.TypePlantingProps{}, pulverization.UseProps{
				Plan: a,
			}),
		)
	})

	e.DELETE("/plan/remove/:id", func(c echo.Context) error {
		id := c.Param("id")
		fmt.Println("Plan removido visualmente:", id)
		return c.NoContent(200)
	})

	e.DELETE("/fertilization/:id", DeleteFertilization)

	e.GET("/plantings", ListPlantings(db))
	e.GET("/plantings/new", ShowPlantingForm(db))
	e.GET("/plantings/edit/:id", ShowPlantingForm(db))
	e.POST("/plantings/create", CreatePlanting(db))
	e.POST("/plantings/update/:id", UpdatePlanting(db))
	e.DELETE("/plantings/delete/:id", DeletePlanting(db))

	e.GET("/plantings/list", ListPlantings(db))

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

	e.GET("/pulverizations", ListPulverizations())

	e.GET("/dashboard/plantings/:planId/pulverization/create", ShowPulverizationForm(db))
	e.POST("/dashboard/plantings/:planId/pulverization/create", CreatePulverization(db))

	// Pulverization
	e.GET("/pulverization", ShowPulverizationForm(db))
	e.GET("/pulverization/:id", ShowPulverizationForm(db))
	e.POST("/pulverization", CreatePulverization(db))
	e.POST("/pulverization/:id", UpdatePulverization(db))
	e.DELETE("/pulverization/:id", DeletePulverization(db))

	// Criação múltipla de pulverizações
	e.GET("/pulverization/multiple", func(c echo.Context) error {
		var as []pulverization.TypePlantingProps

		return pulverization.IndexMult(
			pulverization.PulverizationProps{
				ID:         0,
				PlantingID: 0,
				Unit:       "",
				Products:   []pulverization.ProductInput{},
				Plantings:  []pulverization.TypePlantingProps{},
				Error:      map[string]string{},
				AppliedAt: pulverization.Date{
					Time: time.Now(),
				},
			},
			pulverization.UseProps{
				Prod: []produto.ProductProps{},
			}, as).
			Render(c.Request().Context(), c.Response().Writer)
	})

	e.POST("/pulverization/multiple", CreatePulverizationWithSplit(db))

	// Irrigations
	e.GET("/irrigation/list", IrrigationList)
	e.GET("/irrigation/create", IrrigationCreatePage)
	e.GET("/irrigation/:id", IrrigationShow)
	e.POST("/irrigation/create", IrrigationCreate)
	e.POST("/irrigation/update/:id", IrrigationUpdate)
	e.DELETE("/irrigation/:id", IrrigationDelete)

	// Irrigation Actions
	e.GET("/irrigation-actions", ListIrrigationActions(db))
	e.POST("/irrigation-actions", CreateIrrigationAction(db))
	e.PUT("/irrigation-actions/:id", UpdateIrrigationAction(db))
	e.DELETE("/irrigation-actions/:id", DeleteIrrigationAction(db))

	// Rota HTMX para carregar os detalhes do modal
	e.GET("/irrigation/:id/details", IrrigationDetails)

	e.GET("/dashboard/plantings/:planId/harvest/create", func(c echo.Context) error {
		return harvest.Index(harvest.HarvestProps{
			ID:         0,
			PlantingID: 0,
			HarvestedAt: harvest.Date{
				Time: time.Now(),
			},
			Quantity:  0,
			Unit:      "",
			SaleValue: 0,
			Error:     map[string]string{},
		}).Render(c.Request().Context(), c.Response())
	})
	e.POST("/dashboard/plantings/:planId/harvest/create", CreateHarvest(db))

	// e.GET("/harvest", ListHarvest)
	e.GET("/harvest/:id", ShowHarvest)
	e.GET("/harvest/create", func(c echo.Context) error {
		return harvest.Index(harvest.HarvestProps{}).Render(c.Request().Context(), c.Response())
	})

	e.POST("/harvest/create", CreateHarvest(db))
	e.POST("/harvest/update/:id", UpdateHarvest)
	e.DELETE("/harvest/delete/:id", DeleteHarvest(db))
	// Fertilization routes
	e.GET("/fertilization", ListFertilization)
	e.GET("/fertilization/create", func(c echo.Context) error {
		return fertilization.Index(fertilization.FertilizationProps{
			ID:              0,
			PlantingID:      0,
			ApplicationType: "",
			AppliedAt: fertilization.Date{
				Time: time.Now(),
			},
			Products: []pulverization.ProductInput{},
			Error:    map[string]string{},
		}, pulverization.UseProps{}).
			Render(c.Request().Context(), c.Response())
	})
	e.GET("/dashboard/plantings/:planId/fertilization/create", func(c echo.Context) error {
		return fertilization.Index(fertilization.FertilizationProps{
			ID:              0,
			PlantingID:      0,
			ApplicationType: "",
			AppliedAt: fertilization.Date{
				Time: time.Now(),
			},
			Products: []pulverization.ProductInput{},
			Error:    map[string]string{},
		}, pulverization.UseProps{}).
			Render(c.Request().Context(), c.Response())
	})
	e.GET("/fertilization/create", func(c echo.Context) error {
		return fertilization.Index(fertilization.FertilizationProps{}, pulverization.UseProps{}).
			Render(c.Request().Context(), c.Response())
	})
	e.GET("/fertilization/:id", ShowFertilization)
	e.POST("/dashboard/plantings/:planId/fertilization/create", CreateFertilization)
	e.POST("/fertilization/create", CreateFertilization)
	e.POST("/fertilization/update/:id", UpdateFertilization)

	// Rotas de vendas (sale)
	e.GET("/sales", s.ListSale)          // Lista todas as vendas
	e.GET("/sales/:id", s.ShowSale)      // Mostra detalhes de uma venda
	e.POST("/sales", s.CreateSale)       // Cria uma nova venda
	e.POST("/sales/:id", s.UpdateSale)   // Atualiza uma venda existente
	e.DELETE("/sales/:id", s.DeleteSale) // Deleta uma venda

	e.GET("/newSale", s.NewSale)

	e.GET("/dashboard/plantings/:planId/productsell", ListProductSell)
	e.POST("/dashboard/plantings/:planId/productsell/create", CreateProductSell)

	e.GET("/productsell", ListProductSell)
	e.GET("/productsell/create", CreateViewProductSell)
	e.POST("/productsell/create", CreateProductSell)
	e.GET("/productsell/:id", EditViewProductSell)
	e.POST("/productsell/update/:id", UpdateProductSell)
	e.GET("/", ListDasboard())
	e.GET("/dashboard/plantings/:id", DashboardShowPlanting())
	e.GET("/dashboard/plantings/:id/", DashboardShowPlanting())

	e.POST("/dashboard/plantings/:planId/service/create", CreateService(db))
	e.POST("/service/update/:id", UpdateService(db))
	e.DELETE("/service/delete/:id", DeleteService(db))
	e.GET("/dashboard/plantings/:planId/service", NewService(db))

	// Vacinação (emergencial pós-germinação)
	e.GET("/dashboard/plantings/:planId/vaccination/create", ShowVaccinationForm(db))
	e.POST("/dashboard/plantings/:planId/vaccination/create", CreateVaccination(db))
	e.DELETE("/vaccination/:id", DeleteVaccination(db))
	e.GET("/vaccination/list", ListVaccinations(db))

	// CULTURA ROUTES
	e.GET("/culturas", HandleListCulturas)
	e.GET("/culturas/new", HandleShowCulturaForm)
	e.POST("/culturas/create", HandleCreateCultura)
	e.GET("/culturas/edit/:id", HandleShowCulturaForm)
	e.POST("/culturas/:id", HandleUpdateCultura)
	e.DELETE("/culturas/:id", HandleDeleteCultura)

	// CATEGORIA ROUTES
	e.GET("/categorias", HandleListCategorias)
	e.GET("/categorias/new", HandleShowCategoriaForm)
	e.POST("/categorias/create", HandleCreateCategoria)
	e.GET("/categorias/edit/:id", HandleShowCategoriaForm)
	e.POST("/categorias/:id", HandleUpdateCategoria)
	e.DELETE("/categorias/:id", HandleDeleteCategoria)

	// Pulverização total por área
	e.GET("/pulverization/total-area", func(c echo.Context) error {
		return pulverization.Total(
			pulverization.PulverizationProps{
				ID:         0,
				PlantingID: 0,
				Unit:       "",
				Products:   []pulverization.ProductInput{},
				Error:      map[string]string{},
				AppliedAt: pulverization.Date{
					Time: time.Now(),
				},
			},
			pulverization.UseProps{
				Prod: []produto.ProductProps{},
			},
		).Render(c.Request().Context(), c.Response().Writer)
	})
	e.POST("/pulverization/total-area", CreatePulverizationTotalArea(db))

	// Rotas
	e.GET("/typeProduct", s.ListTypeProduct)
	e.GET("/listTypeProduct", s.ListTypeProduct)
	e.GET("/typeProduct/create", func(c echo.Context) error {
		// Nova instância vazia
		props := typeproduct.TypeProductProps{
			Error: map[string]string{},
		}
		return Render(c, 200, typeproduct.Index(props))
	})

	e.GET("/typeProduct/:ID", s.EditTypeProduct)
	e.POST("/typeProduct/create", s.CreateTypeProduct)
	e.POST("/typeProduct/update/:ID", s.UpdateTypeProduct)
	e.POST("/typeProduct/delete/:ID", s.DeleteTypeProduct)

	e.GET("/cashflow", ListCashFlows)

	e.GET("/cashflow/create", ShowCreateCashFlow)
	e.POST("/cashflow/create", CreateCashFlow)
	e.GET("/cashflow/:id", ShowCashFlow)
	e.POST("/cashflow/update/:id", UpdateCashFlow)
	e.DELETE("/cashflow/:id", DeleteCashFlow)

	e.Logger.Fatal(e.Start(":1323"))
}

func seedCategorias() {
	categorias := []string{
		"Acaricida",
		"Adjuvante",
		"Bactericida",
		"Espalhante Adesivo",
		"Estimulante",
		"Fertilizante",
		"Fungicida",
		"Herbicida",
		"Inoculante",
		"Inseticida",
		"Nematicida",
		"Óleo Mineral",
		"Óleo Vegetal",
		"Regulador de Crescimento",
		"Surfactante",
	}
	for _, name := range categorias {
		var existing Categoria
		if err := db.Where("LOWER(name) = LOWER(?)", name).First(&existing).Error; err != nil {
			db.Create(&Categoria{Name: name})
		}
	}
}

func nullableDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}
