package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/isaquerr25/go-templ-htmx/views/pages/irrigation"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

var (
	irrigations []irrigation.IrrigationProps
	nextID      uint = 1
)

var irrigationStore = []irrigation.IrrigationProps{} // simulação de DB na memória

// GET /irrigation
func IrrigationList(c echo.Context) error {
	var irrigations []Irrigation // model do banco

	if err := db.Find(&irrigations).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Erro ao buscar irrigações")
	}

	// Converte para props da view, se necessário
	items := make([]irrigation.IrrigationProps, 0, len(irrigations))
	for _, i := range irrigations {
		items = append(items, irrigation.IrrigationProps{
			ID:         i.ID,
			PlantingID: i.PlantingID,
			StartedAt:  i.AppliedAt,
			// ... outros campos se houver
		})
	}

	data := irrigation.IrrigationListProps{
		Items: items,
	}

	return irrigation.List(data).Render(c.Request().Context(), c.Response().Writer)
}

// GET /irrigation/:id
func IrrigationShow(c echo.Context) error {
	// Converte o ID
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		return c.NoContent(http.StatusBadRequest)
	}

	// Busca do banco
	var model Irrigation
	if err := db.First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.NoContent(http.StatusNotFound)
		}
		fmt.Println("Erro ao buscar irrigação:", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	fmt.Println("model.Method:", model.Method)

	// Conversão para props esperada pela view
	props := irrigation.IrrigationProps{
		ID:         model.ID,
		PlantingID: model.PlantingID,
		Type:       model.Method,
		StartedAt:  model.AppliedAt,
		Error:      make(map[string]string),
	}

	fmt.Println("props:", props.Type)

	// Pega os plantios (ex: para dropdown)
	plan, _ := GetAllPlantings()

	// Renderiza com os dados convertidos
	return irrigation.Index(props, plan).
		Render(c.Request().Context(), c.Response().Writer)
}

// GET /irrigation/create
func IrrigationCreatePage(c echo.Context) error {
	plan, _ := GetAllPlantings()
	return irrigation.Index(irrigation.IrrigationProps{}, plan).
		Render(c.Request().Context(), c.Response().Writer)
}

// POST /irrigation/create
func IrrigationCreate(c echo.Context) error {
	var input irrigation.IrrigationProps
	input.Error = make(map[string]string)

	// Validação: PlantingID
	plantingID, err := strconv.Atoi(c.FormValue("plantingId"))
	if err != nil || plantingID <= 0 {
		input.Error["PlantingID"] = "ID do plantio inválido"
	} else {
		input.PlantingID = uint(plantingID)
	}

	// Validação: StartedAt
	startedAtStr := c.FormValue("startedAt")
	startedAt, err := time.Parse("2006-01-02T15:04", startedAtStr)
	if err != nil {
		input.Error["StartedAt"] = "Data de início inválida"
	} else {
		input.StartedAt = startedAt
	}

	// Buscar todos os plantios, para exibir se der erro
	plan, _ := GetAllPlantings()

	// Se houver erros, renderiza com eles
	if len(input.Error) > 0 {
		return irrigation.Index(input, plan).Render(c.Request().Context(), c.Response())
	}

	// Salvar no banco usando GORM
	irrigationModel := Irrigation{
		PlantingID: input.PlantingID,
		AppliedAt:  input.StartedAt,
		Method:     input.Type,
		Notes:      "",
	}
	if err := db.Create(&irrigationModel).Error; err != nil {
		fmt.Println("Erro ao salvar irrigação:", err)
		return err
	}

	// Redireciona se tudo deu certo
	c.Response().Header().Set("HX-Redirect", "/irrigation/list")
	return c.NoContent(http.StatusOK)
}

// POST /irrigation/update/:id
func IrrigationUpdate(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		return c.String(http.StatusBadRequest, "ID inválido")
	}

	// Buscar do banco
	var model Irrigation
	if err := db.First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.String(http.StatusNotFound, "Registro não encontrado")
		}
		fmt.Println("Erro ao buscar irrigação:", err)
		return c.String(http.StatusInternalServerError, "Erro interno")
	}

	// Recebe os dados do form
	var input irrigation.IrrigationProps
	input.Error = make(map[string]string)

	// PlantingID
	plantingID, err := strconv.Atoi(c.FormValue("plantingId"))
	if err != nil || plantingID <= 0 {
		input.Error["PlantingID"] = "ID do plantio inválido"
	} else {
		input.PlantingID = uint(plantingID)
	}

	// Type
	input.Type = c.FormValue("type")
	if input.Type == "" {
		input.Error["Type"] = "Tipo de irrigação obrigatório"
	}

	// StartedAt
	startedAtStr := c.FormValue("startedAt")
	startedAt, err := time.Parse("2006-01-02T15:04", startedAtStr)
	if err != nil {
		input.Error["StartedAt"] = "Data de início inválida"
	} else {
		input.StartedAt = startedAt
	}

	input.ID = uint(id)

	// Em caso de erro de validação, renderiza novamente o formulário
	if len(input.Error) > 0 {
		plan, _ := GetAllPlantings()
		fmt.Println("Erro ao atualizar irrigação:", input.Error)

		return irrigation.Index(input, plan).Render(c.Request().Context(), c.Response().Writer)
	}

	fmt.Println(" Method", input.Type)

	// Atualiza o modelo
	model.PlantingID = input.PlantingID
	model.Method = input.Type
	model.AppliedAt = input.StartedAt

	// Salva no banco
	if err := db.Save(&model).Error; err != nil {
		fmt.Println("Erro ao atualizar irrigação:", err)
		return c.String(http.StatusInternalServerError, "Erro ao salvar")
	}

	// Redireciona para a lista
	c.Response().Header().Set("HX-Redirect", "/irrigation/list")
	return c.NoContent(http.StatusOK)
}

// DELETE /irrigation/:id
func IrrigationDelete(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	for i, item := range irrigationStore {
		if int(item.ID) == id {
			irrigationStore = append(irrigationStore[:i], irrigationStore[i+1:]...)
			break
		}
	}
	return c.NoContent(http.StatusOK)
}

func ListIrrigationActions(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var actions []IrrigationAction
		if err := db.Find(&actions).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao buscar ações")
		}
		return c.JSON(http.StatusOK, actions)
	}
}

func CreateIrrigationAction(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var input IrrigationAction

		if err := c.Bind(&input); err != nil {
			return c.String(http.StatusBadRequest, "Erro ao ler dados")
		}

		if input.AppliedAt.IsZero() {
			input.AppliedAt = time.Now()
		}

		if err := db.Create(&input).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao salvar ação")
		}

		return c.JSON(http.StatusCreated, input)
	}
}

func UpdateIrrigationAction(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))

		var action IrrigationAction
		if err := db.First(&action, id).Error; err != nil {
			return c.String(http.StatusNotFound, "Ação não encontrada")
		}

		if err := c.Bind(&action); err != nil {
			return c.String(http.StatusBadRequest, "Erro ao ler dados")
		}

		if err := db.Save(&action).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao atualizar ação")
		}

		return c.JSON(http.StatusOK, action)
	}
}

func DeleteIrrigationAction(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))

		if err := db.Delete(&IrrigationAction{}, id).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao deletar ação")
		}

		return c.NoContent(http.StatusNoContent)
	}
}

func IrrigationDetails(c echo.Context) error {
	id := c.Param("id")

	// Aqui futuramente você pode buscar do banco de dados os detalhes da irrigação
	// Por enquanto, vamos simular com HTML estático

	html := fmt.Sprintf(`
		<ul class="list-disc pl-5">
			<li>Setor 1 – 10 minutos</li>
			<li>Setor 2 – 15 minutos</li>
			<li>Setor 3 – 20 minutos</li>
		</ul>
		<p class="mt-2 text-sm text-gray-400">Irrigação vinculada ao ID: %s</p>
	`, id)

	return c.HTML(http.StatusOK, html)
}
