package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/isaquerr25/go-templ-htmx/views/pages/produto"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Server struct{}

func GetAllProductsProps() ([]produto.ProductProps, error) {
	var products []Product
	if err := db.Preload("Lots").Find(&products).Error; err != nil {
		return nil, err
	}

	var list []produto.ProductProps
	for _, p := range products {
		var totalQty, totalRemaining, totalCost float64
		for _, lot := range p.Lots {
			totalQty += lot.Quantity
			totalRemaining += lot.Remaining
			totalCost += lot.TotalCost
		}

		list = append(list, produto.ProductProps{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Unit:        p.Unit,
			BaseValue:   p.BaseValue,
			Quantity:    totalQty,
			Remaining:   totalRemaining,
			TotalCost:   totalCost,
			Error:       map[string]string{},
		})
	}
	return list, nil
}

func GetAllProductsForUserProps() ([]produto.ProductProps, error) {
	var products []Product
	if err := db.Preload("Lots", func(db *gorm.DB) *gorm.DB {
		return db.Where("remaining > 0").Order("date ASC")
	}).Preload("Categorias.Categoria").Find(&products).Error; err != nil {
		return nil, err
	}

	var list []produto.ProductProps
	for _, p := range products {
		var totalRemaining float64
		for _, lot := range p.Lots {
			totalRemaining += lot.Remaining
		}
		if totalRemaining <= 0 {
			continue
		}

		var catItems []produto.CategoriaCheckItem
		catLabels := ""
		for i, pc := range p.Categorias {
			catItems = append(catItems, produto.CategoriaCheckItem{
				ID:   pc.CategoriaID,
				Name: pc.Categoria.Name,
			})
			if i > 0 {
				catLabels += ", "
			}
			catLabels += pc.Categoria.Name
		}

		list = append(list, produto.ProductProps{
			ID:                   p.ID,
			Name:                 p.Name,
			Description:          p.Description,
			Remaining:            totalRemaining,
			Unit:                 p.Unit,
			PrePulverizationBase: p.PrePulverizationBase,
			AllCategorias:        catItems,
			CategoriasLabel:      catLabels,
			Error:                map[string]string{},
		})
	}
	return list, nil
}

func consumeStock(tx *gorm.DB, productID uint, quantity float64) error {
	var lots []ProductLot
	if err := tx.Where("product_id = ? AND remaining > 0", productID).Order("date ASC, id ASC").Find(&lots).Error; err != nil {
		return err
	}

	toConsume := quantity
	for _, lot := range lots {
		if toConsume <= 0 {
			break
		}
		if lot.Remaining >= toConsume {
			lot.Remaining -= toConsume
			toConsume = 0
		} else {
			toConsume -= lot.Remaining
			lot.Remaining = 0
		}
		if err := tx.Save(&lot).Error; err != nil {
			return err
		}
	}

	if toConsume > 0 {
		return fmt.Errorf("estoque insuficiente")
	}
	return nil
}

func restoreStock(tx *gorm.DB, productID uint, quantity float64) error {
	var lots []ProductLot
	if err := tx.Where("product_id = ?", productID).Order("date DESC, id DESC").Find(&lots).Error; err != nil {
		return err
	}

	toRestore := quantity
	for _, lot := range lots {
		if toRestore <= 0 {
			break
		}
		origQty := lot.Quantity
		canRestore := origQty - lot.Remaining
		if canRestore > toRestore {
			canRestore = toRestore
		}
		lot.Remaining += canRestore
		toRestore -= canRestore
		if err := tx.Save(&lot).Error; err != nil {
			return err
		}
	}
	return nil
}

// ========================
// LIST PRODUCTS
// ========================
func HandleListProduct(c echo.Context) error {
	nameFilter := strings.TrimSpace(c.QueryParam("name"))
	categoriaFilter := strings.TrimSpace(c.QueryParam("categoria"))

	var products []Product
	query := db.Where("1 = 1").Preload("Lots", func(db *gorm.DB) *gorm.DB {
		return db.Order("date DESC")
	}).Preload("Categorias")

	if nameFilter != "" {
		query = query.Where("LOWER(name) LIKE LOWER(?)", "%"+nameFilter+"%")
	}

	if err := query.Order("name ASC").Find(&products).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Erro ao buscar produtos")
	}

	var listProps []produto.ProductListProps
	for _, product := range products {
		var totalQuantity, totalRemaining, totalCost float64
		for _, lot := range product.Lots {
			totalQuantity += lot.Quantity
			totalRemaining += lot.Remaining
			totalCost += lot.TotalCost
		}

		catNames := getLinkedCategoriaNames(product.ID)
		listProps = append(listProps, produto.ProductListProps{
			ID:          product.ID,
			ProductID:   product.ID,
			Name:        product.Name,
			Description: product.Description,
			Quantity:    totalQuantity,
			Remaining:   totalRemaining,
			TotalCost:   totalCost,
			Categorias:  catNames,
			Error:       map[string]string{},
		})
	}

	// Filter by categoria if selected
	if categoriaFilter != "" {
		var filtered []produto.ProductListProps
		for _, p := range listProps {
			var pcs []ProductCategoria
			db.Where("product_id = ? AND categoria_id = ?", p.ProductID, categoriaFilter).Find(&pcs)
			if len(pcs) > 0 {
				filtered = append(filtered, p)
			}
		}
		listProps = filtered
	}

	allCategorias := getAllCategorias()

	if c.Request().Header.Get("HX-Request") == "true" {
		return Render(c, http.StatusOK, produto.ListContent(listProps, nameFilter, allCategorias, categoriaFilter))
	}
	return Render(c, http.StatusOK, produto.List(listProps, nameFilter, allCategorias, categoriaFilter))
}

// ========================
// SHOW PRODUCT
// ========================
func HandleShowProduct(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("ID"))
	if err != nil {
		return c.String(http.StatusBadRequest, "ID inválido")
	}

	var product Product
	if err := db.Preload("Lots", func(db *gorm.DB) *gorm.DB {
		return db.Order("date DESC")
	}).First(&product, id).Error; err != nil {
		return c.String(http.StatusNotFound, "Produto não encontrado")
	}

	var lotsProps []produto.LotProps
	var totalQuantity, totalRemaining, totalCost float64

	for _, lot := range product.Lots {
		totalQuantity += lot.Quantity
		totalRemaining += lot.Remaining
		totalCost += lot.TotalCost

		lotsProps = append(lotsProps, produto.LotProps{
			ID:        lot.ID,
			Quantity:  lot.Quantity,
			Remaining: lot.Remaining,
			UnitCost:  lot.UnitCost,
			TotalCost: lot.TotalCost,
			Date:      lot.Date,
		})
	}

	unit := product.Unit
	if unit == "" && len(product.Lots) > 0 {
		unit = product.Lots[0].Unit
	}

	cultures := getLinkedCulturas(product.ID)
	categorias := getAllCategoriasWithSelection(product.ID)

	props := produto.ProductProps{
		ID:            product.ID,
		Name:          product.Name,
		Description:   product.Description,
		BaseValue:     product.BaseValue,
		Unit:          unit,
		Lots:          lotsProps,
		Cultures:      cultures,
		AllCategorias: categorias,
		Quantity:      totalQuantity,
		Remaining:     totalRemaining,
		TotalCost:     totalCost,
		Error:         map[string]string{},
	}

	return Render(c, http.StatusOK, produto.Index(props))
}

// ========================
// NEW PRODUCT FORM
// ========================
func HandleNewProduct(c echo.Context) error {
	return Render(c, 200, produto.Index(produto.ProductProps{
		ID:            0,
		Name:          "",
		Lots:          []produto.LotProps{},
		AllCategorias: getAllCategorias(),
		Error:         map[string]string{},
		IsNew:         true,
	}))
}

// ========================
// CREATE PRODUCT
// ========================
func HandleCreateProduct(c echo.Context) error {
	name := strings.TrimSpace(c.FormValue("name"))
	description := strings.TrimSpace(c.FormValue("description"))
	unit := strings.TrimSpace(c.FormValue("unit"))
	baseValueStr := strings.TrimSpace(c.FormValue("baseValue"))

	errorsMap := make(map[string]string)

	if name == "" {
		errorsMap["Name"] = "O nome do produto é obrigatório"
	}

	if unit == "" {
		errorsMap["Unit"] = "A unidade é obrigatória"
	}

	var baseValue float64
	if baseValueStr != "" {
		baseValue, _ = strconv.ParseFloat(strings.ReplaceAll(baseValueStr, ",", "."), 64)
	}

	categoriaIDs := c.Request().Form["categoria_ids"]
	selectedCategorias := buildSelectedCategorias(categoriaIDs)

	if len(errorsMap) > 0 {
		return Render(c, 200, produto.Index(produto.ProductProps{
			Name:          name,
			Unit:          unit,
			IsNew:         true,
			Error:         errorsMap,
			Lots:          []produto.LotProps{},
			AllCategorias: selectedCategorias,
		}))
	}

	product := Product{
		Name:        name,
		Description: description,
		Unit:        unit,
		BaseValue:   baseValue,
	}

	if err := db.Create(&product).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Erro ao criar produto")
	}

	saveProductCategorias(db, product.ID, categoriaIDs)

	c.Response().Header().Set("HX-Redirect", "/listProduct")
	return c.NoContent(http.StatusOK)
}

// ========================
// UPDATE PRODUCT
// ========================
func HandleUpdateProduct(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("ID"))
	if err != nil {
		return c.String(http.StatusBadRequest, "ID inválido")
	}

	var product Product
	if err := db.First(&product, id).Error; err != nil {
		return c.String(http.StatusNotFound, "Produto não encontrado")
	}

	product.Name = strings.TrimSpace(c.FormValue("name"))
	product.Description = strings.TrimSpace(c.FormValue("description"))
	product.Unit = strings.TrimSpace(c.FormValue("unit"))

	baseValueStr := strings.TrimSpace(c.FormValue("baseValue"))
	if baseValueStr != "" {
		v, _ := strconv.ParseFloat(strings.ReplaceAll(baseValueStr, ",", "."), 64)
		product.BaseValue = v
	}

	errorsMap := make(map[string]string)
	if product.Name == "" {
		errorsMap["Name"] = "O nome não pode ser vazio"
	}

	categoriaIDs := c.Request().Form["categoria_ids"]

	if len(errorsMap) == 0 {
		if err := db.Save(&product).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao atualizar produto")
		}
		saveProductCategorias(db, product.ID, categoriaIDs)
		c.Response().Header().Set("HX-Redirect", "/listProduct")
		return c.NoContent(http.StatusOK)
	}

	var lotsProps []produto.LotProps
	for _, lot := range product.Lots {
		lotsProps = append(lotsProps, produto.LotProps{
			ID:        lot.ID,
			Quantity:  lot.Quantity,
			Remaining: lot.Remaining,
			UnitCost:  lot.UnitCost,
			TotalCost: lot.TotalCost,
			Date:      lot.Date,
		})
	}

	props := produto.ProductProps{
		ID:            product.ID,
		Name:          product.Name,
		Unit:          product.Unit,
		Lots:          lotsProps,
		AllCategorias: buildSelectedCategorias(categoriaIDs),
		Error:         errorsMap,
	}

	return Render(c, 200, produto.Index(props))
}

// ========================
// EDIT PRODUCT
// ========================
func HandleEditProduct(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("ID"))
	if err != nil {
		return c.String(http.StatusBadRequest, "ID inválido")
	}

	var product Product
	if err := db.Preload("Lots").First(&product, id).Error; err != nil {
		return c.String(http.StatusNotFound, "Produto não encontrado")
	}

	var lotsProps []produto.LotProps
	for _, lot := range product.Lots {
		lotsProps = append(lotsProps, produto.LotProps{
			ID:        lot.ID,
			Quantity:  lot.Quantity,
			Remaining: lot.Remaining,
			UnitCost:  lot.UnitCost,
			TotalCost: lot.TotalCost,
			Date:      lot.Date,
		})
	}

	cultures := getLinkedCulturas(product.ID)
	categorias := getAllCategoriasWithSelection(product.ID)

	props := produto.ProductProps{
		ID:            product.ID,
		Name:          product.Name,
		Description:   product.Description,
		Unit:          product.Unit,
		BaseValue:     product.BaseValue,
		Lots:          lotsProps,
		Cultures:      cultures,
		AllCategorias: categorias,
		Error:         map[string]string{},
	}

	return Render(c, 200, produto.Index(props))
}

// ========================
// DELETE PRODUCT
// ========================
func HandleDeleteProduct(c echo.Context) error {
	id := c.Param("ID")

	result := db.Delete(&Product{}, id)
	if result.Error != nil {
		return c.String(http.StatusInternalServerError, "Erro ao deletar produto")
	}
	if result.RowsAffected == 0 {
		return c.String(http.StatusNotFound, "Produto não encontrado")
	}

	c.Response().Header().Set("HX-Refresh", "true")
	return c.NoContent(http.StatusNoContent)
}

func HandleRemoveProductVisual(c echo.Context) error {
	id := c.Param("id")
	fmt.Println("Produto removido visualmente:", id)
	return c.NoContent(200)
}

// ========================
// ADD LOT MODAL
// ========================
func HandleAddLot(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "ID inválido")
	}

	props := produto.LotProps{
		ProductID: uint(id),
		Date:      time.Now(),
	}

	return Render(c, http.StatusOK, produto.CreateLotModal(props))
}

// ========================
// CREATE LOT
// ========================
func HandleCreateLot(c echo.Context) error {
	productParam := c.Param("id")
	productID, err := strconv.Atoi(productParam)
	if err != nil {
		return c.String(http.StatusBadRequest, "Produto inválido")
	}

	var product Product
	if err := db.First(&product, productID).Error; err != nil {
		return c.String(http.StatusNotFound, "Produto não encontrado")
	}

	parseF := func(v string) float64 {
		f, _ := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(v), ",", "."), 64)
		return f
	}

	quantity := parseF(c.FormValue("quantity"))
	if quantity <= 0 {
		return c.String(http.StatusBadRequest, "Quantidade deve ser maior que zero")
	}

	totalCost := parseF(c.FormValue("totalCost"))
	if totalCost < 0 {
		return c.String(http.StatusBadRequest, "Valor total inválido")
	}

	remaining := parseF(c.FormValue("remaining"))
	if remaining < 0 || remaining > quantity {
		return c.String(http.StatusBadRequest, "Restante inválido")
	}

	unit := strings.TrimSpace(c.FormValue("unit"))
	if unit == "" {
		unit = product.Unit
	}

	purchaseDateStr := strings.TrimSpace(c.FormValue("purchaseDate"))
	var purchaseDate time.Time
	if purchaseDateStr != "" {
		purchaseDate, err = time.ParseInLocation("2006-01-02", purchaseDateStr, time.Local)
		if err != nil {
			return c.String(http.StatusBadRequest, "Data inválida")
		}
	} else {
		purchaseDate = time.Now()
	}

	unitCost := 0.0
	if quantity > 0 {
		unitCost = totalCost / quantity
	}

	lot := ProductLot{
		ProductID: uint(productID),
		Quantity:  quantity,
		Remaining: remaining,
		Unit:      unit,
		UnitCost:  unitCost,
		TotalCost: totalCost,
		Date:      purchaseDate,
	}

	if err := db.Create(&lot).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Erro ao salvar lote")
	}

	props := produto.LotProps{
		ID:        lot.ID,
		ProductID: lot.ProductID,
		Quantity:  lot.Quantity,
		Remaining: lot.Remaining,
		UnitCost:  lot.UnitCost,
		TotalCost: lot.TotalCost,
		Date:      lot.Date,
	}

	return Render(c, http.StatusOK, produto.LotRow(props))
}

// ========================
// EDIT LOT MODAL
// ========================
func HandleEditLot(c echo.Context) error {
	lotID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "ID do lote inválido")
	}

	var lot ProductLot
	if err := db.First(&lot, lotID).Error; err != nil {
		return c.String(http.StatusNotFound, "Lote não encontrado")
	}

	return Render(c, http.StatusOK, produto.EditLotModal(produto.LotProps{
		ID:        lot.ID,
		Quantity:  lot.Quantity,
		Remaining: lot.Remaining,
		UnitCost:  lot.UnitCost,
		TotalCost: lot.TotalCost,
		Date:      lot.Date,
	}))
}

// ========================
// UPDATE LOT
// ========================
func HandleUpdateLot(c echo.Context) error {
	lotID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "ID inválido")
	}

	var lot ProductLot
	if err := db.First(&lot, lotID).Error; err != nil {
		return c.String(http.StatusNotFound, "Lote não encontrado")
	}

	parseF := func(v string) float64 {
		f, _ := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(v), ",", "."), 64)
		return f
	}

	qty := parseF(c.FormValue("quantity"))
	rem := parseF(c.FormValue("remaining"))
	totalCost := parseF(c.FormValue("totalCost"))
	dateStr := strings.TrimSpace(c.FormValue("purchaseDate"))

	if qty <= 0 {
		return c.String(http.StatusUnprocessableEntity, "Quantidade deve ser maior que zero")
	}
	if rem < 0 {
		return c.String(http.StatusUnprocessableEntity, "Restante não pode ser negativo")
	}
	if rem > qty {
		return c.String(http.StatusUnprocessableEntity, "Restante não pode ser maior que o total")
	}

	unitCost := 0.0
	if totalCost > 0 {
		unitCost = totalCost / qty
	}

	var purchaseDate time.Time
	if dateStr != "" {
		purchaseDate, err = time.ParseInLocation("2006-01-02", dateStr, time.Local)
		if err != nil {
			return c.String(http.StatusBadRequest, "Data inválida")
		}
	} else {
		purchaseDate = lot.Date
	}

	err = db.Model(&lot).Updates(map[string]interface{}{
		"quantity":   qty,
		"remaining":  rem,
		"unit_cost":  unitCost,
		"total_cost": totalCost,
		"date":       purchaseDate,
	}).Error
	if err != nil {
		return c.String(http.StatusInternalServerError, "Erro ao salvar no banco")
	}

	props := produto.LotProps{
		ID:        lot.ID,
		ProductID: lot.ProductID,
		Quantity:  qty,
		Remaining: rem,
		UnitCost:  unitCost,
		TotalCost: totalCost,
		Date:      purchaseDate,
	}

	return Render(c, http.StatusOK, produto.LotRow(props))
}

// ========================
// DELETE LOT
// ========================
func HandleDeleteLot(c echo.Context) error {
	lotID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Lote inválido")
	}

	if err := db.Delete(&ProductLot{}, lotID).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Erro ao deletar lote")
	}

	return c.NoContent(http.StatusOK)
}

// ========================
// CULTURAS VINCULADAS
// ========================
func getLinkedCulturas(productID uint) []produto.CultureRow {
	var pcs []ProductCultura
	db.Where("product_id = ?", productID).Preload("Cultura").Find(&pcs)

	var cultures []produto.CultureRow
	for _, pc := range pcs {
		culturaName := ""
		if pc.Cultura.ID != 0 {
			culturaName = pc.Cultura.Name
		}
		cultures = append(cultures, produto.CultureRow{
			PCID:        pc.ID,
			CulturaID:   pc.CulturaID,
			CulturaName: culturaName,
			Proportion:  pc.Proportion,
		})
	}
	return cultures
}

func getLinkedCategoriaNames(productID uint) []string {
	var pcs []ProductCategoria
	db.Where("product_id = ?", productID).Preload("Categoria").Find(&pcs)
	var names []string
	for _, pc := range pcs {
		if pc.Categoria.ID != 0 {
			names = append(names, pc.Categoria.Name)
		}
	}
	return names
}

func getAllCategoriasWithSelection(productID uint) []produto.CategoriaCheckItem {
	var all []Categoria
	db.Order("name ASC").Find(&all)

	var selectedIDs []uint
	var pcs []ProductCategoria
	db.Where("product_id = ?", productID).Find(&pcs)
	for _, pc := range pcs {
		selectedIDs = append(selectedIDs, pc.CategoriaID)
	}

	selectedMap := make(map[uint]bool)
	for _, id := range selectedIDs {
		selectedMap[id] = true
	}

	var items []produto.CategoriaCheckItem
	for _, cat := range all {
		items = append(items, produto.CategoriaCheckItem{
			ID:       cat.ID,
			Name:     cat.Name,
			Selected: selectedMap[cat.ID],
		})
	}
	return items
}

func buildSelectedCategorias(ids []string) []produto.CategoriaCheckItem {
	all := getAllCategorias()
	selectedMap := make(map[uint]bool)
	for _, idStr := range ids {
		id, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err == nil {
			selectedMap[uint(id)] = true
		}
	}
	for i := range all {
		if selectedMap[all[i].ID] {
			all[i].Selected = true
		}
	}
	return all
}

func getAllCategorias() []produto.CategoriaCheckItem {
	var all []Categoria
	db.Order("name ASC").Find(&all)
	var items []produto.CategoriaCheckItem
	for _, cat := range all {
		items = append(items, produto.CategoriaCheckItem{
			ID: cat.ID,
			Name: cat.Name,
		})
	}
	return items
}

func saveProductCategorias(tx *gorm.DB, productID uint, categoriaIDs []string) error {
	tx.Where("product_id = ?", productID).Delete(&ProductCategoria{})
	for _, idStr := range categoriaIDs {
		id, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err != nil || id <= 0 {
			continue
		}
		if err := tx.Create(&ProductCategoria{
			ProductID:   productID,
			CategoriaID: uint(id),
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func getAllCulturasForEdit(productID uint) []produto.CultureRow {
	linked := getLinkedCulturas(productID)

	linkedIDs := map[uint]bool{}
	for _, c := range linked {
		linkedIDs[c.CulturaID] = true
	}

	var allDb []Cultura
	db.Order("name ASC").Find(&allDb)

	var rows []produto.CultureRow
	rows = append(rows, linked...)
	for _, ct := range allDb {
		if !linkedIDs[ct.ID] {
			rows = append(rows, produto.CultureRow{
				CulturaID:   ct.ID,
				CulturaName: ct.Name,
			})
		}
	}
	return rows
}
