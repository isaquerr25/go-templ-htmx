package main

import (
	"time"

	"gorm.io/gorm"
)

// Produto representa um item no banco de dados
type Product struct {
	gorm.Model
	Name        string    `form:"name"`        // Product name
	Quantity    float64   `form:"quantity"`    // Total purchased quantity
	Remaining   float64   `form:"remaining"`   // Remaining in stock
	Unit        string    `form:"unit"`        // e.g., kg, L
	Date        time.Time `form:"date"`        // Purchase date
	TotalCost   float64   `form:"totalCost"`   // Total cost for the batch
	Description string    `form:"description"` // Notes or comments
}

type Field struct {
	gorm.Model
	Name        string  `form:"name"`
	Hectares    float64 `form:"hectares"`
	Description string  `form:"description"`
}

type Planting struct {
	gorm.Model
	TypeProductID *uint      `form:"typeProductId"`
	CropName      string     `form:"cropName"`
	StartedAt     time.Time  `form:"startedAt"`
	EndedAt       *time.Time `form:"endedAt"`
	IsCompleted   bool       `form:"isCompleted"`
	AreaUsed      float64    `form:"areaUsed"`
}

type Fertilization struct {
	gorm.Model
	PlantingID      uint                 `form:"plantingId"`
	ApplicationType string               `form:"applicationType"` // drip or foliar
	AppliedAt       time.Time            `form:"appliedAt"`
	Products        []ApplyFertilization `form:"foreignKey:FertilizationID"`
}

type ApplyFertilization struct {
	gorm.Model
	FertilizationID uint    `form:"fertilizationId"`
	ProductID       uint    `form:"productId"`
	QuantityUsed    float64 `form:"quantityUsed"`
	Unit            string  `form:"unit"`
}

type Pulverization struct {
	gorm.Model
	PlantingID uint             `form:"plantingId"`
	AppliedAt  time.Time        `form:"appliedAt"`
	Unit       string           `form:"unit"`
	Products   []AppliedProduct `form:"foreignKey:PulverizationID"`
}

type AppliedProduct struct {
	gorm.Model
	PulverizationID uint    `form:"pulverizationId"`
	ProductID       uint    `form:"productId"`
	QuantityUsed    float64 `form:"quantityUsed"`
}

type Irrigation struct {
	gorm.Model
	PlantingID uint      `form:"plantingId"`
	AppliedAt  time.Time `form:"appliedAt"`
	Method     string    `form:"method"` // drip, sprinkler, etc.
	Notes      string    `form:"notes"`
}

type IrrigationAction struct {
	gorm.Model
	IrrigationID uint      `form:"plantingId"`
	AppliedAt    time.Time `form:"appliedAt"`
}

type Harvest struct {
	gorm.Model
	PlantingID  uint      `form:"plantingId"`
	HarvestedAt time.Time `form:"harvestedAt"`
	Quantity    float64   `form:"quantity"` // Amount harvested
	Unit        string    `form:"unit"`     // e.g., kg, liter
}

// Client representa um cliente (comprador ou parceiro)
type Client struct {
	gorm.Model
	Name    string `form:"name"`
	Email   string `form:"email"`
	Phone   string `form:"phone"`
	Company string `form:"company"`
	Address string `form:"address"`
	Notes   string `form:"notes"`
}

// ProductSell representa um produto disponível para venda
type ProductSell struct {
	gorm.Model
	Name        string  `form:"name"`
	Description string  `form:"description"`
	Unit        string  `form:"unit"`  // kg, L, etc.
	Price       float64 `form:"price"` // Preço por unidade
	Stock       float64 `form:"stock"` // Quantidade disponível
}

// Sale representa uma venda feita para um cliente
type Sale struct {
	gorm.Model
	ClientID      *uint     `form:"clientId"`
	ProductSellID uint      `form:"productSellId"`
	SoldAt        time.Time `form:"soldAt"`

	Loss bool `form:"loss"`

	Quantity   float64 `form:"quantity"`
	Unit       string  `form:"unit"`       // Repetido para facilitar acesso direto
	TotalPrice float64 `form:"totalPrice"` // Total da venda

	Method SaleMethod `form:"method"` // ex: dinheiro, cartão, pix
	State  SaleState  `form:"state"`  // ex: pendente, pago, cancelado

	Notes string `form:"notes"`
}

type Service struct {
	gorm.Model
	Name        string    `form:"name"`        // Nome do serviço (ex: Transporte, Consultoria)
	Description string    `form:"description"` // Descrição adicional
	Cost        float64   `form:"cost"`        // Custo total do serviço
	PlantingID  *uint     `form:"plantingId"`  // Opcional, caso o serviço tenha sido aplicado a um plantio
	Notes       string    `form:"notes"`       // Observações específicas
	CreateAt    time.Time `form:"performedAt"`
}

type TypeProduct struct {
	gorm.Model
	Name     string  `form:"name"`
	Describe string  `form:"describe"`
	Quantity float64 `form:"quantity"` // Custo total do serviço
}

type (
	SaleMethod string
	SaleState  string
)

const (
	SaleStatePending   SaleState = "pending"
	SaleStateCompleted SaleState = "completed"
	SaleStateCanceled  SaleState = "canceled"

	SaleMethodCash   SaleMethod = "cash"
	SaleMethodPix    SaleMethod = "pix"
	SaleMethodCard   SaleMethod = "card"
	SaleMethodCredit SaleMethod = "credit"
)

func regraDeTres(a, b, c float64) float64 {
	if a == 0 {
		panic("Divisão por zero: 'a' não pode ser 0")
	}
	return (b * c) / a
}
