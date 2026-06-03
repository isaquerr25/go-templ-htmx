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

func computeCurrentStageFromCultura(startedAt time.Time, c Cultura) string {
	days := int(time.Since(startedAt).Hours() / 24)
	if days < 0 {
		days = 0
	}
	if c.MorteInicio > 0 && days >= c.MorteInicio {
		return "morte"
	}
	if c.ColheitaInicio > 0 && days >= c.ColheitaInicio {
		return "colheita"
	}
	if c.FloracaoInicio > 0 && days >= c.FloracaoInicio {
		return "floracao"
	}
	if c.GerminacaoInicio > 0 && days >= c.GerminacaoInicio {
		return "germinacao"
	}
	return "plantio"
}

func computeCurrentStage(startedAt time.Time, tp TypeProduct) string {
	days := int(time.Since(startedAt).Hours() / 24)
	if days < 0 {
		days = 0
	}
	if tp.MorteInicio > 0 && days >= tp.MorteInicio {
		return "morte"
	}
	if tp.ColheitaInicio > 0 && days >= tp.ColheitaInicio {
		return "colheita"
	}
	if tp.FloracaoInicio > 0 && days >= tp.FloracaoInicio {
		return "floracao"
	}
	if tp.GerminacaoInicio > 0 && days >= tp.GerminacaoInicio {
		return "germinacao"
	}
	return "plantio"
}

func stageLabel(stage string) string {
	switch stage {
	case "plantio":
		return "🌱 Plantio"
	case "germinacao":
		return "🌿 Germinação"
	case "floracao":
		return "🌸 Floração"
	case "colheita":
		return "🌾 Colheita"
	case "morte":
		return "🍂 Morte"
	}
	return stage
}

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

		stage := ""
		if p.CulturaID != nil {
			var ct Cultura
			if err := db.First(&ct, *p.CulturaID).Error; err == nil {
				stage = stageLabel(computeCurrentStageFromCultura(p.StartedAt, ct))
			}
		}

		culturaID := uint(0)
		if p.CulturaID != nil {
			culturaID = *p.CulturaID
		}

		plantings = append(plantings, planting.PlantingProps{
			ID:          p.ID,
			CulturaID:   culturaID,
			CropName:    p.CropName,
			StartedAt:   p.StartedAt.Format("2006-01-02"),
			EndedAt:     endedAtStr,
			IsCompleted: p.IsCompleted,
			AreaUsed:    p.AreaUsed,
			CurrentStage: stage,
			Error:       nil,
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

	culturaId, _ := strconv.ParseFloat(c.FormValue("culturaId"), 64)
	// Preenche props
	props.CropName = cropName
	props.StartedAt = startedAtStr
	props.EndedAt = endedAtStr
	props.IsCompleted = isCompleted
	props.AreaUsed = areaUsed
	props.CulturaID = uint(culturaId)

	fmt.Printf("Props validados: %+v\n", props)

	return props, hasError, nil
}

func ListPlantings(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		q := strings.TrimSpace(c.QueryParam("q"))
		completed := c.QueryParam("completed")
		culturaIDStr := c.QueryParam("culturaId")

		var plantings []Planting
		query := db.Model(&Planting{})

		if q != "" {
			query = query.Where("crop_name LIKE ?", "%"+q+"%")
		}
		if completed == "true" {
			query = query.Where("is_completed = ?", true)
		} else if completed == "false" {
			query = query.Where("is_completed = ?", false)
		}
		if culturaIDStr != "" {
			if id, err := strconv.Atoi(culturaIDStr); err == nil {
				query = query.Where("cultura_id = ?", id)
			}
		}

		if err := query.Find(&plantings).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao buscar plantios")
		}

		var allCulturas []planting.FilterCultura
		var ctList []Cultura
		db.Find(&ctList)
		for _, ct := range ctList {
			allCulturas = append(allCulturas, planting.FilterCultura{
				ID:   ct.ID,
				Name: ct.Name,
			})
		}

		var items []planting.PlantingItem
		for _, p := range plantings {
			stage := ""
			culturaName := ""
			if p.CulturaID != nil {
				var ct Cultura
				if err := db.First(&ct, *p.CulturaID).Error; err == nil {
					stage = stageLabel(computeCurrentStageFromCultura(p.StartedAt, ct))
					culturaName = ct.Name
				}
			}

			items = append(items, planting.PlantingItem{
				ID:           p.ID,
				CulturaID:    p.CulturaID,
				CropName:     p.CropName,
				StartedAt:    p.StartedAt,
				EndedAt:      p.EndedAt,
				IsCompleted:  p.IsCompleted,
				AreaUsed:     p.AreaUsed,
				CurrentStage: stage,
				CulturaName:  culturaName,
			})
		}

		return planting.List(items, allCulturas, q, completed, culturaIDStr).Render(c.Request().Context(), c.Response().Writer)
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
				planting.Index(props, []planting.CulturaProps{}),
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

		culturaID := &props.CulturaID
		if props.CulturaID == 0 {
			culturaID = nil
		}

		newPlanting := Planting{
			CropName:      props.CropName,
			StartedAt:     startedAt,
			EndedAt:       endedAt,
			IsCompleted:   props.IsCompleted,
			AreaUsed:      props.AreaUsed,
			CulturaID:     culturaID,
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
				planting.Index(props, []planting.CulturaProps{}),
			)
		}

		startedAt, err := time.Parse("2006-01-02", props.StartedAt)
		if err != nil {
			props.Error = map[string]string{"StartedAt": "Data inválida"}
			return c.Render(
				http.StatusOK,
				"main",
				planting.Index(props, []planting.CulturaProps{}),
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
					planting.Index(props, []planting.CulturaProps{}),
				)
			}
			endedAt = &t
		}

		var plant Planting
		if err := db.First(&plant, id).Error; err != nil {
			return c.String(http.StatusNotFound, "Plantio não encontrado")
		}

		culturaID := &props.CulturaID
		if props.CulturaID == 0 {
			culturaID = nil
		}

		plant.CropName = props.CropName
		plant.StartedAt = startedAt
		plant.EndedAt = endedAt
		plant.IsCompleted = props.IsCompleted
		plant.AreaUsed = props.AreaUsed
		plant.CulturaID = culturaID

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
		var culturas []Cultura

		if err := db.Find(&culturas).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Erro ao buscar culturas")
		}

		props := make([]planting.CulturaProps, len(culturas))
		for i, ct := range culturas {
			props[i] = planting.CulturaProps{
				ID:   ct.ID,
				Name: ct.Name,
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

		var culturaID uint
		if plant.CulturaID != nil {
			culturaID = *plant.CulturaID
		}

		p := planting.PlantingProps{
			ID:         plant.ID,
			CropName:   plant.CropName,
			StartedAt:  plant.StartedAt.Format("2006-01-02"),
			EndedAt:    endedAt,
			IsCompleted: plant.IsCompleted,
			AreaUsed:   plant.AreaUsed,
			Error:      map[string]string{},
			CulturaID:  culturaID,
		}

		return planting.Index(p, props).
			Render(c.Request().Context(), c.Response().Writer)
	}
}
