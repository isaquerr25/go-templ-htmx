package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/isaquerr25/go-templ-htmx/views/pages/planting"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func GetAllPlantings() ([]planting.PlantingProps, error) {
	var dbPlantings []Planting
	if err := db.Find(&dbPlantings).Error; err != nil {
		return nil, err
	}

	var plantings []planting.PlantingProps
	for _, p := range dbPlantings {
		var endedAtStr string
		if p.EndedAt != nil {
			endedAtStr = p.EndedAt.Format("2006-01-02")
		}

		plantings = append(plantings, planting.PlantingProps{
			ID:          p.ID,
			CropName:    p.CropName,
			StartedAt:   p.StartedAt.Format("2006-01-02"),
			EndedAt:     endedAtStr,
			IsCompleted: p.IsCompleted,
			AreaUsed:    p.AreaUsed,
			Error:       nil, // pode ser preenchido posteriormente em outra lógica
		})
	}

	return plantings, nil
}

func validatePlanting(c echo.Context) (props planting.PlantingProps, hasError bool, err error) {
	props.Error = map[string]string{}

	// Captura dos campos do formulário
	fieldIDStr := strings.TrimSpace(c.FormValue("fieldId"))
	cropName := strings.TrimSpace(c.FormValue("cropName"))
	startedAtStr := strings.TrimSpace(c.FormValue("startedAt"))
	endedAtStr := strings.TrimSpace(c.FormValue("endedAt"))
	isCompletedStr := c.FormValue("isCompleted")
	areaUsedStr := strings.TrimSpace(c.FormValue("areaUsed"))

	fmt.Printf(
		"Validando plantio: fieldId=%q, cropName=%q, startedAt=%q, endedAt=%q, isCompleted=%q, areaUsed=%q\n",
		fieldIDStr,
		cropName,
		startedAtStr,
		endedAtStr,
		isCompletedStr,
		areaUsedStr,
	)

	// Validações
	if cropName == "" {
		props.Error["CropName"] = "Nome da cultura é obrigatório"
		hasError = true
		fmt.Println("Erro de validação: Nome da cultura vazio")
	}

	if startedAtStr == "" {
		props.Error["StartedAt"] = "Data de início é obrigatória"
		hasError = true
		fmt.Println("Erro de validação: Data de início vazia")
	} else if _, errParse := time.Parse("2006-01-02", startedAtStr); errParse != nil {
		props.Error["StartedAt"] = "Data de início inválida (formato: AAAA-MM-DD)"
		hasError = true
		fmt.Printf("Erro ao converter StartedAt: %v\n", errParse)
	}

	if endedAtStr != "" {
		if _, errParse := time.Parse("2006-01-02", endedAtStr); errParse != nil {
			props.Error["EndedAt"] = "Data final inválida (formato: AAAA-MM-DD)"
			hasError = true
			fmt.Printf("Erro ao converter EndedAt: %v\n", errParse)
		}
	}

	var areaUsed float64
	var errParseArea error

	if areaUsedStr == "" {
		props.Error["AreaUsed"] = "Área usada é obrigatória"
		hasError = true
		fmt.Println("Erro de validação: Área usada vazia")
	} else {
		areaUsed, errParseArea = strconv.ParseFloat(areaUsedStr, 64)
		if errParseArea != nil {
			props.Error["AreaUsed"] = "Área usada inválida"
			hasError = true
			fmt.Printf("Erro ao converter AreaUsed: %v\n", errParseArea)
		}
	}

	// Interpretação do checkbox
	isCompleted := isCompletedStr == "on" || isCompletedStr == "true" || isCompletedStr == "1"

	typeProdutcId, _ := strconv.ParseFloat(c.FormValue("typeProdutcId"), 64)
	// Preenche props
	props.CropName = cropName
	props.StartedAt = startedAtStr
	props.EndedAt = endedAtStr
	props.IsCompleted = isCompleted
	props.AreaUsed = areaUsed
	props.TypePoductID = uint(typeProdutcId)

	fmt.Printf("Props validados: %+v\n", props)

	return props, hasError, nil
}

func ListPlantings(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var plantings []Planting

		if err := db.Find(&plantings).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao buscar plantios")
		}

		var items []planting.PlantingItem
		for _, p := range plantings {
			items = append(items, planting.PlantingItem{
				ID:          p.ID,
				CropName:    p.CropName,
				StartedAt:   p.StartedAt,
				EndedAt:     p.EndedAt,
				IsCompleted: p.IsCompleted,
				AreaUsed:    p.AreaUsed,
			})
		}

		// Gerar HTML via templ do go-templ-htmx
		return planting.List(items).Render(c.Request().Context(), c.Response().Writer)
		// Responder com HTML
	}
}

func CreatePlanting(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		props, hasError, err := validatePlanting(c)
		if err != nil {
			fmt.Printf("Erro ao validar plantio: %v\n", err)
			return c.String(
				http.StatusBadRequest,
				"Erro técnico ao processar dados do formulário: "+err.Error(),
			)
		}
		if hasError {
			fmt.Println("Erro de validação no formulário, renderizando página novamente.")
			return c.Render(
				http.StatusOK,
				"main",
				planting.Index(props, []planting.TypeProductProps{}),
			)
		}

		// Conversão final para time.Time
		startedAt, err := time.Parse("2006-01-02", props.StartedAt)
		if err != nil {
			fmt.Printf("Erro ao converter StartedAt: %v\n", err)
			return c.String(http.StatusBadRequest, "Data de início inválida")
		}

		var endedAt *time.Time
		if props.EndedAt != "" {
			t, err := time.Parse("2006-01-02", props.EndedAt)
			if err != nil {
				fmt.Printf("Erro ao converter EndedAt: %v\n", err)
				return c.String(http.StatusBadRequest, "Data de término inválida")
			}
			endedAt = &t
		}

		newPlanting := Planting{
			CropName:      props.CropName,
			StartedAt:     startedAt,
			EndedAt:       endedAt,
			IsCompleted:   props.IsCompleted,
			AreaUsed:      props.AreaUsed,
			TypeProductID: &props.TypePoductID,
		}

		if err := db.Create(&newPlanting).Error; err != nil {
			fmt.Printf("Erro ao salvar no banco de dados: %v\n", err)
			return c.String(
				http.StatusInternalServerError,
				"Erro ao salvar plantio no banco de dados: "+err.Error(),
			)
		}

		c.Response().Header().Set("HX-Redirect", "/")
		return c.String(http.StatusOK, "")
	}
}

func UpdatePlanting(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.String(http.StatusBadRequest, "ID inválido")
		}

		// Usa a mesma função de validação do create para ler os dados do form e validar
		props, hasError, err := validatePlanting(c)
		if err != nil {
			fmt.Printf("Erro ao validar plantio: %v\n", err)
			return c.String(
				http.StatusBadRequest,
				"Erro técnico ao processar dados do formulário: "+err.Error(),
			)
		}

		if hasError {
			fmt.Println("Erro de validação no formulário, renderizando página novamente.")
			return c.Render(
				http.StatusOK,
				"main",
				planting.Index(props, []planting.TypeProductProps{}),
			)
		}

		startedAt, err := time.Parse("2006-01-02", props.StartedAt)
		if err != nil {
			props.Error = map[string]string{"StartedAt": "Data inválida"}
			return c.Render(
				http.StatusOK,
				"main",
				planting.Index(props, []planting.TypeProductProps{}),
			)
		}

		var endedAt *time.Time
		if props.EndedAt != "" {
			t, err := time.Parse("2006-01-02", props.EndedAt)
			if err != nil {
				props.Error = map[string]string{"EndedAt": "Data final inválida"}
				return c.Render(
					http.StatusOK,
					"main",
					planting.Index(props, []planting.TypeProductProps{}),
				)
			}
			endedAt = &t
		}

		var plant Planting
		if err := db.First(&plant, id).Error; err != nil {
			return c.String(http.StatusNotFound, "Plantio não encontrado")
		}

		plant.CropName = props.CropName
		plant.StartedAt = startedAt
		plant.EndedAt = endedAt
		plant.IsCompleted = props.IsCompleted
		plant.AreaUsed = props.AreaUsed
		plant.TypeProductID = &props.TypePoductID // atenção ao typo aqui

		if err := db.Save(&plant).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao atualizar plantio")
		}

		c.Response().Header().Set("HX-Redirect", "/")
		return c.String(http.StatusOK, "")
	}
}

func DeletePlanting(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))

		if err := db.Delete(&Planting{}, id).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao deletar plantio")
		}

		c.Response().Header().Set("HX-Redirect", "/")
		return c.String(http.StatusOK, "")
	}
}

func ShowPlantingForm(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var pp []TypeProduct

		if err := db.Find(&pp).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao buscar tipos de produtos")
		}

		props := make([]planting.TypeProductProps, len(pp))
		for i, p := range pp {
			props[i] = planting.TypeProductProps{
				ID:   p.ID,
				Name: p.Name,
			}
		}

		id := c.Param("id")
		if id == "" {
			// Novo cadastro
			p := planting.PlantingProps{
				ID:          0,
				CropName:    "milho",
				StartedAt:   time.Now().Format("2006-01-02"),
				AreaUsed:    10.0,
				IsCompleted: false,
				EndedAt:     "",
				Error:       map[string]string{},
			}
			return planting.Index(p, props).
				Render(c.Request().Context(), c.Response().Writer)
		}

		// Edição
		plantID, err := strconv.Atoi(id)
		if err != nil {
			return c.String(http.StatusBadRequest, "ID inválido")
		}

		var plant Planting
		if err := db.First(&plant, plantID).Error; err != nil {
			return c.String(http.StatusNotFound, "Plantio não encontrado")
		}

		var endedAt string
		if plant.EndedAt != nil {
			endedAt = plant.EndedAt.Format("2006-01-02")
		}

		var typeProductID uint
		if plant.TypeProductID != nil {
			typeProductID = *plant.TypeProductID
		}

		p := planting.PlantingProps{
			ID:           plant.ID,
			CropName:     plant.CropName,
			StartedAt:    plant.StartedAt.Format("2006-01-02"),
			EndedAt:      endedAt,
			IsCompleted:  plant.IsCompleted,
			AreaUsed:     plant.AreaUsed,
			Error:        map[string]string{},
			TypePoductID: typeProductID,
		}

		return planting.Index(p, props).
			Render(c.Request().Context(), c.Response().Writer)
	}
}
