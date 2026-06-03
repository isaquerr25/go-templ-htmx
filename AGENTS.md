# go-templ-htmx

Go + Echo v4 + GORM + SQLite + templ + HTMX + Tailwind CSS — agricultural management system focused on planting lifecycle (fields, plantings, fertilizations, pulverizations, irrigations, harvests, vaccinations, services, cash flow, sales).

## Commands

| Command | Action |
|---|---|
| `make build` | Build binary to `./dist/server` |
| `make run` | Build + run server |
| `make dev` | Run all 4 watchers in parallel (templ generate --watch, air, tailwind --watch, asset sync) |
| `templ generate` | Regenerate Go code from `.templ` files |
| `npx tailwindcss -i ./views/input.css -o ./static/vendor/tailwind.css --minify` | Build Tailwind |

## Architecture

- **All `.go` files are `package main`** in the repo root — no `internal/` or `cmd/`.
- **Global `var db *gorm.DB`** — most handlers capture it via closures; `Server` struct (in `product.go`) used only for Product/Client/Sale/TypeProduct handlers.
- **SQLite** via `gorm.io/driver/sqlite` — `base.db` checked into VC. No PostgreSQL.
- **templ codegen**: each `views/**/*.templ` produces sibling `_templ.go` (compiled Go) and `_templ.txt` (watch marker). Generated files checked in.
- **HTMX v2** drives frontend partial rendering via `hx-*` attributes. No session/auth middleware.
- **Tailwind CSS** vendored binary in root; content scans `./views/**/*.templ`.
- **No authentication, no multi-tenancy, no organizations** — single-user mode.
- All views under `views/pages/<entity>/` with `Index`, `List`, `Show`, `Form` patterns.
- Single base template at `views/templates/index.templ`.

## Data Models (`models.go`)

| Model | Purpose | Key Relations |
|---|---|---|
| `Product` | Input materials (agrochemicals, fertilizers) | HasMany `ProductLot` (FIFO stock tracking) |
| `ProductLot` | Individual purchase lots with remaining quantity | BelongsTo `Product` |
| `ProductRecommendation` | Recommended product + dose per crop type | ProductID + TypeProductID unique |
| `Field` | Physical field/area | Standalone |
| `Planting` | Crop planting cycle | Optional TypeProductID, HasMany Fertilization/Pulverization/Irrigation/Harvest/Service/Vaccination |
| `TypeProduct` | Crop type definition with phenological stages (germination/flowering/harvest/death days) | Referenced by Plantings |
| `Fertilization` | Fertilizer application event | HasMany `ApplyFertilization`, BelongsTo Planting |
| `ApplyFertilization` | Product used in fertilization | BelongsTo Fertilization |
| `Pulverization` | Pesticide application event | HasMany `AppliedProduct`, BelongsTo Planting |
| `AppliedProduct` | Product consumed in pulverization | BelongsTo Pulverization, consumes stock via FIFO |
| `Irrigation` | Irrigation event | BelongsTo Planting |
| `IrrigationAction` | Sub-actions within irrigation | BelongsTo Irrigation |
| `Harvest` | Harvest event (updates TypeProduct.Quantity) | BelongsTo Planting |
| `Vaccination` | Post-germination vaccination | BelongsTo Planting |
| `Service` | External service with cost tracking | Optional PlantingID |
| `Client` | Customer/buyer | Standalone |
| `ProductSell` | Product available for sale (output goods) | Standalone |
| `Sale` | Sale transaction | ClientID, ProductSellID, Loss flag |
| `CashFlow` | Financial transaction (in/out) | Optional SaleID/ServiceID/ClientID refs |
| `SaleMethod`/`SaleState` | Enums for payment method and sale status | |
| `FlowType`/`FlowMethod`/`FlowCategory` | Enums for cash flow classification | |

## Handler Files

| File | Functions | Style |
|---|---|---|
| `main.go` | Entry point, route registration, `Render()` helper, `validateProduct()` | Inline handlers + closure pattern |
| `product.go` | `Server` struct, CRUD, `consumeStock()`/`restoreStock()` FIFO helpers | Struct methods |
| `client.go` | `Server` CRUD methods | Struct methods |
| `sale.go` | `Server` CRUD methods | Struct methods |
| `typeproduct.go` | `Server` CRUD + `validateTypeProduct()` | Struct methods |
| `planting.go` | CRUD + `computeCurrentStage()`, `stageLabel()`, `GetAllPlantings()` | Closure pattern with `*gorm.DB` param |
| `field.go` | `Field` (physical field) CRUD + `GetAllFields()` (returns TypeProduct props — confusing) | Closure + standalone |
| `pulverization.go` | CRUD + `CreatePulverizationWithSplit()` (distribute % across plantings) + `CreatePulverizationTotalArea()` (distribute by area) | Closure with `*gorm.DB` param |
| `fertilization.go` | CRUD with FIFO stock consumption/restore | Standalone functions |
| `irrigation.go` | CRUD + `IrrigationActions` sub-resource + `IrrigationDetails` | Mixed (standalone + closure) |
| `harvest.go` | CRUD — updates `TypeProduct.Quantity` on create/delete | Closure with `*gorm.DB` param |
| `vaccination.go` | CRUD with planting name lookup | Closure with `*gorm.DB` param |
| `service.go` | CRUD | Closure with `*gorm.DB` param |
| `cashflow.go` | CRUD with search/filter + in/out totals | Standalone functions |
| `recommendation.go` | CRUD — JSON API for product-crop recommendations | Closure with `*gorm.DB` param |
| `dashboard.go` | `ListDasboard()` (filterable planting list) + `DashboardShowPlanting()` (cost/harvest summary) | Closure returning `echo.HandlerFunc` |

## Key Patterns

- **FIFO stock**: `consumeStock()` consumes oldest lots first; `restoreStock()` returns to recent lots. Used by pulverization and fertilization.
- **`regraDeTres()`** (rule of three): `(b * c) / a` — calculates proportional cost from total qty/cost.
- **`HX-Redirect`** header for HTMX navigation after form submission — no JSON responses.
- **No validation packages** — validation inline in each handler/bind.
- **`Server` struct** (empty) serves as method receiver for Product/Client/Sale/TypeProduct routes only.
- **Planting stage computation** based on `TypeProduct` phenological day offsets since `StartedAt`.

## Quirks

- **Confusing naming**: `Field` struct exists but `GetAllFields()` returns `TypeProduct` data; `field.go` handler manages physical `Field` model but helper mixes types.
- **`typeProductId` typo**: Planting form field named `typeProdutcId` (missing 'c') — consistent across validation and model.
- **Irrigation DELETE**: uses in-memory `irrigationStore` slice instead of DB.
- **`CreatePulverization`** divides raw quantity by 1000 (assumes input in mL, stores in L).
- **No cascade deletes** on Planting — child records orphaned on delete.
- **`Sale` model** saves `SoldAt` as string via `time.Time.String()` instead of `time.Time` — breaks date filtering.
- **`Harvest`** modifies `TypeProduct.Quantity` (which is supposed to be a crop type template, not inventory).
- **No auth, no sessions** — single-user mode, no middleware.
- **`go vet` / `go build`** is the only validation available — no linters or tests.
