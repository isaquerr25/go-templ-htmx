package main

import (
	"time"

	"gorm.io/gorm"
)

// Produto representa um item no banco de dados
type Product struct {
	gorm.Model
	Name                 string    `form:"name"`
	Quantity             float64   `form:"quantity"`
	Remaining            float64   `form:"remaining"`
	Unit                 string    `form:"unit"`
	Date                 time.Time `form:"date"`
	TotalCost            float64   `form:"totalCost"`
	Description          string    `form:"description"`
	PrePulverizationBase float64   `form:"prePulverizationBase"` // base de pré-pulverização
}

type Field struct {
	gorm.Model
	Name        string  `form:"name"`
	Hectares    float64 `form:"hectares"`
	Description string  `form:"description"`
}

type Planting struct {
	gorm.Model
	TypeProductID *uint      `form:"typeProductId" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	CropName      string     `form:"cropName"`
	StartedAt     time.Time  `form:"startedAt"`
	EndedAt       *time.Time `form:"endedAt"`
	IsCompleted   bool       `form:"isCompleted"`
	AreaUsed      float64    `form:"areaUsed"`
}

type Fertilization struct {
	gorm.Model
	PlantingID      uint                 `form:"plantingId"                 gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ApplicationType string               `form:"applicationType"`
	AppliedAt       time.Time            `form:"appliedAt"`
	Products        []ApplyFertilization `form:"foreignKey:FertilizationID" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type ApplyFertilization struct {
	gorm.Model
	FertilizationID uint    `form:"fertilizationId" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ProductID       uint    `form:"productId"`
	QuantityUsed    float64 `form:"quantityUsed"`
	Unit            string  `form:"unit"`
}

type Pulverization struct {
	gorm.Model
	PlantingID uint             `form:"plantingId"                 gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	AppliedAt  time.Time        `form:"appliedAt"`
	Unit       string           `form:"unit"`
	Products   []AppliedProduct `form:"foreignKey:PulverizationID" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type AppliedProduct struct {
	gorm.Model
	PulverizationID uint    `form:"pulverizationId" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ProductID       uint    `form:"productId"`
	QuantityUsed    float64 `form:"quantityUsed"`
}

type Irrigation struct {
	gorm.Model
	PlantingID uint      `form:"plantingId" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	AppliedAt  time.Time `form:"appliedAt"`
	Method     string    `form:"method"`
	Notes      string    `form:"notes"`
}

type IrrigationAction struct {
	gorm.Model
	IrrigationID uint      `form:"plantingId" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	AppliedAt    time.Time `form:"appliedAt"`
}

type Harvest struct {
	gorm.Model
	PlantingID  uint      `form:"plantingId"  gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	HarvestedAt time.Time `form:"harvestedAt"`
	Quantity    float64   `form:"quantity"`
	Unit        string    `form:"unit"`
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
	Unit        string  `form:"unit"`
	Price       float64 `form:"price"`
	Stock       float64 `form:"stock"`
}

// Sale representa uma venda feita para um cliente
type Sale struct {
	gorm.Model
	ClientID      *uint     `form:"clientId"      gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	ProductSellID uint      `form:"productSellId" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	SoldAt        time.Time `form:"soldAt"`

	Loss bool `form:"loss"`

	Quantity   float64 `form:"quantity"`
	Unit       string  `form:"unit"`
	TotalPrice float64 `form:"totalPrice"`

	Method SaleMethod `form:"method"`
	State  SaleState  `form:"state"`

	Notes string `form:"notes"`
}

type Service struct {
	gorm.Model
	Name        string    `form:"name"`
	Description string    `form:"description"`
	Cost        float64   `form:"cost"`
	PlantingID  *uint     `form:"plantingId"  gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Notes       string    `form:"notes"`
	CreateAt    time.Time `form:"performedAt"`
}

type TypeProduct struct {
	gorm.Model
	Name     string  `form:"name"`
	Describe string  `form:"describe"`
	Quantity float64 `form:"quantity"`
}

// CashFlow representa uma movimentação financeira (entrada ou saída de dinheiro)
type CashFlow struct {
	gorm.Model
	Type        FlowType     `form:"type"`        // entrada ou saída
	Category    FlowCategory `form:"category"`    // ex: venda, compra, serviço, despesa, outro
	Amount      float64      `form:"amount"`      // valor da transação
	Method      FlowMethod   `form:"method"`      // pix, dinheiro, cartão, transferência...
	OccurredAt  time.Time    `form:"occurredAt"`  // data da transação
	Description string       `form:"description"` // descrição livre

	// Referências opcionais para relacionar com outras tabelas
	SaleID    *uint `form:"saleId"    gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	ServiceID *uint `form:"serviceId" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	ClientID  *uint `form:"clientId"  gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`

	Notes string `form:"notes"`
}

// FlowType define se é entrada ou saída
type FlowType string

const (
	FlowTypeIn  FlowType = "in"  // entrada
	FlowTypeOut FlowType = "out" // saída
)

// FlowMethod define o método de pagamento ou recebimento
type FlowMethod string

const (
	FlowMethodCash        FlowMethod = "cash"        // dinheiro
	FlowMethodPix         FlowMethod = "pix"         // pix
	FlowMethodCard        FlowMethod = "card"        // cartão
	FlowMethodTransfer    FlowMethod = "transfer"    // transferência bancária
	FlowMethodInstallment FlowMethod = "installment" // a prazo
)

// FlowCategory define o tipo de origem/destino do dinheiro
type FlowCategory string

const (
	FlowCategorySale       FlowCategory = "sale"       // entrada por venda
	FlowCategoryService    FlowCategory = "service"    // entrada por serviço
	FlowCategoryPurchase   FlowCategory = "purchase"   // saída por compra
	FlowCategoryExpense    FlowCategory = "expense"    // saída por despesa geral
	FlowCategoryInvestment FlowCategory = "investment" // saída/entrada de investimento
	FlowCategoryOther      FlowCategory = "other"      // outro tipo
)

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
