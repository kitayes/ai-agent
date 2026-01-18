package llm

import (
	"fmt"
	"regexp"
	"strings"

	"arcgis-ai-assistant/internal/models"
)

// BuildPromptWithContext creates an enhanced prompt with project context
func BuildPromptWithContext(userRequest string, context *models.Context) string {
	systemPrompt := `Ты автономный ГИС-инженер и эксперт по ArcGIS Python API (ArcPy).

ТВОЯ РОЛЬ:
Ты — интеллектуальный агент, который превращает сложные ГИС-системы в среду, работающую на основе естественного языка. Ты не просто генерируешь код — ты ПРОЕКТИРУЕШЬ и ПЛАНИРУЕШЬ решения геоинформационных задач, анализируешь пространственные связи и самостоятельно принимаешь решения о том, как достичь цели.

ПРАВИЛА ГЕНЕРАЦИИ КОДА:
1. Генерируй ТОЛЬКО безопасный Python код с использованием arcpy
2. Используй ТОЛЬКО слои, которые доступны в проекте (список ниже)
3. Если пользователь просит слой, которого нет — предложи ближайший по смыслу
4. Код должен быть готов к немедленному выполнению
5. Используй arcpy.AddMessage() для вывода результатов пользователю
6. Обрабатывай ошибки через try-except где необходимо
7. Для сложных операций разбивай на шаги с сообщениями о прогрессе

ЗАПРЕЩЕНО:
- os.remove, shutil.rmtree (удаление файлов)
- subprocess, os.system (системные команды)
- open() для записи (кроме временных файлов arcpy)
- urllib, requests (сетевые запросы)`

	if context != nil && len(context.Layers) > 0 {
		systemPrompt += "\n\n" + formatContextInfo(context)
	} else {
		systemPrompt += "\n\nВНИМАНИЕ: Контекст проекта недоступен. Используй общие подходы."
	}

	systemPrompt += `

ФОРМАТ ОТВЕТА:
` + "```python" + `
import arcpy

# Твой код здесь
# Используй arcpy.AddMessage() для вывода информации

` + "```" + `

ОБЪЯСНЕНИЕ: Краткое описание того, что делает код (на русском языке)`

	return fmt.Sprintf("%s\n\nЗАПРОС ПОЛЬЗОВАТЕЛЯ: %s", systemPrompt, userRequest)
}

// BuildRegenerationPrompt creates a prompt for error correction
func BuildRegenerationPrompt(originalPrompt, failedCode, errorMessage string, context *models.Context, attempt int) string {
	prompt := fmt.Sprintf(`Ты автономный ГИС-инженер. Твой предыдущий код вызвал ошибку. Проанализируй и исправь.

ОРИГИНАЛЬНЫЙ ЗАПРОС: %s

ПОПЫТКА: %d/3

КОД, КОТОРЫЙ НЕ СРАБОТАЛ:
`+"```python"+`
%s
`+"```"+`

ОШИБКА:
%s

`, originalPrompt, attempt, failedCode, errorMessage)

	if context != nil {
		prompt += formatContextInfo(context) + "\n\n"
	}

	prompt += `ЗАДАЧА:
1. Проанализируй ошибку
2. Определи причину (неправильное имя слоя? неверный синтаксис? логическая ошибка?)
3. Сгенерируй ИСПРАВЛЕННЫЙ код

ВАЖНО:
- Используй точные имена слоев из контекста
- Проверь синтаксис ArcPy
- Добавь проверки на существование данных если нужно

ФОРМАТ ОТВЕТА:
` + "```python" + `
# Исправленный код
` + "```" + `

ОБЪЯСНЕНИЕ: Что было исправлено и почему`

	return prompt
}

// formatContextInfo formats context into readable text
func formatContextInfo(context *models.Context) string {
	var sb strings.Builder

	sb.WriteString("═══════════════════════════════════════════\n")
	sb.WriteString("ДОСТУПНЫЕ ДАННЫЕ В ПРОЕКТЕ:\n")
	sb.WriteString("═══════════════════════════════════════════\n\n")

	sb.WriteString(fmt.Sprintf("📁 ПРОЕКТ: %s\n", context.Project.Name))
	sb.WriteString(fmt.Sprintf("📍 Система координат: %s\n\n", context.Project.SpatialReference))

	sb.WriteString(fmt.Sprintf("📊 СЛОИ (%d):\n", len(context.Layers)))

	for i, layer := range context.Layers {
		sb.WriteString(fmt.Sprintf("\n%d. \"%s\"", i+1, layer.Name))

		if layer.GeometryType != "" {
			sb.WriteString(fmt.Sprintf("\n   Тип: %s", layer.GeometryType))
		}

		sb.WriteString(fmt.Sprintf("\n   Объектов: %d", layer.FeatureCount))

		if len(layer.Fields) > 0 {
			sb.WriteString("\n   Поля: ")
			fieldNames := make([]string, 0, len(layer.Fields))
			for _, field := range layer.Fields {
				fieldNames = append(fieldNames, field.Name)
			}
			sb.WriteString(strings.Join(fieldNames, ", "))
		}

		if layer.IsVisible {
			sb.WriteString(" [видимый]")
		}
	}

	if context.ActiveLayer != "" {
		sb.WriteString(fmt.Sprintf("\n\n🎯 АКТИВНЫЙ СЛОЙ: \"%s\"\n", context.ActiveLayer))
	}

	sb.WriteString("\n═══════════════════════════════════════════\n")

	return sb.String()
}

// BuildPrompt creates a simple prompt without context (legacy support)
func BuildPrompt(userRequest string) string {
	systemPrompt := `Ты эксперт по ArcGIS Python API (ArcPy). Твоя задача - генерировать безопасный и корректный Python код для ArcGIS Pro.

ПРАВИЛА:
1. Генерируй ТОЛЬКО код на Python с использованием модуля arcpy
2. Код должен быть безопасным (без удаления файлов, системных команд)
3. Используй arcpy.AddMessage() для вывода сообщений пользователю
4. Код должен быть готов к выполнению без дополнительных изменений
5. Добавь краткое объяснение на русском языке

ФОРМАТ ОТВЕТА:
` + "```python" + `
# твой код здесь
` + "```" + `

ОБЪЯСНЕНИЕ: краткое описание того, что делает код`

	return fmt.Sprintf("%s\n\nЗАПРОС ПОЛЬЗОВАТЕЛЯ: %s", systemPrompt, userRequest)
}

func ExtractCodeAndExplanation(response string) (string, string) {
	codePattern := regexp.MustCompile("(?s)```python\\s*\n(.*?)```")
	matches := codePattern.FindStringSubmatch(response)

	code := ""
	if len(matches) > 1 {
		code = strings.TrimSpace(matches[1])
	}

	explanationPattern := regexp.MustCompile("(?i)ОБЪЯСНЕНИЕ:\\s*(.+)")
	expMatches := explanationPattern.FindStringSubmatch(response)

	explanation := ""
	if len(expMatches) > 1 {
		explanation = strings.TrimSpace(expMatches[1])
	} else {
		explanation = "Код сгенерирован успешно"
	}

	if code == "" {
		lines := strings.Split(response, "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "arcpy.") {
				code = trimmed
				break
			}
		}
	}

	return code, explanation
}

// BuildVisionPrompt creates a prompt for vision analysis
func BuildVisionPrompt(userRequest string, context *models.Context) string {
	prompt := `Ты автономный ГИС-аналитик и эксперт по пространственному анализу.

ТВОЯ ЗАДАЧА: Проанализировать скриншот карты ArcGIS и помочь пользователю.

ИНСТРУКЦИИ ПО АНАЛИЗУ:
1. Внимательно изучи карту - что на ней изображено?
2. Определи тип данных (точки, линии, полигоны)
3. Обрати внимание на пространственные паттерны (кластеры, дыры в данных, выбросы)
4. Оцени качество визуализации
5. Предложи конкретные действия для решения задачи пользователя

`

	if context != nil && len(context.Layers) > 0 {
		prompt += formatContextInfo(context) + "\n\n"
	}

	prompt += fmt.Sprintf("ВОПРОС ПОЛЬЗОВАТЕЛЯ: %s\n\n", userRequest)

	prompt += `ФОРМАТ ОТВЕТА:

АНАЛИЗ: Детальное описание того, что ты видишь на карте (2-3 предложения)

ПРЕДЛОЖЕНИЯ:
- Конкретное действие 1
- Конкретное действие 2
- Конкретное действие 3

ЕСЛИ НУЖНО, СГЕНЕРИРУЙ КОД:
` + "```python" + `
# Код для выполнения предложенных действий (если применимо)
` + "```" + `

ОБЪЯСНЕНИЕ: Краткое пояснение кода`

	return prompt
}

// ParseVisionResponse parses the response from Gemini Vision
func ParseVisionResponse(response string) (analysis string, suggestedActions []string, code, explanation string) {
	// Extract analysis
	analysisPattern := regexp.MustCompile(`(?i)АНАЛИЗ:\s*([^\n]+(?:\n(?!ПРЕДЛОЖЕНИЯ:)[^\n]+)*)`)
	analysisMatches := analysisPattern.FindStringSubmatch(response)
	if len(analysisMatches) > 1 {
		analysis = strings.TrimSpace(analysisMatches[1])
	}

	// Extract suggested actions
	suggestionsPattern := regexp.MustCompile(`(?i)ПРЕДЛОЖЕНИЯ:\s*((?:[-•*]\s*[^\n]+\n?)+)`)
	suggestionsMatches := suggestionsPattern.FindStringSubmatch(response)
	if len(suggestionsMatches) > 1 {
		suggestionsText := suggestionsMatches[1]
		// Split by lines and extract actions
		lines := strings.Split(suggestionsText, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			// Remove bullet points
			line = strings.TrimPrefix(line, "-")
			line = strings.TrimPrefix(line, "•")
			line = strings.TrimPrefix(line, "*")
			line = strings.TrimSpace(line)
			if line != "" {
				suggestedActions = append(suggestedActions, line)
			}
		}
	}

	// Extract code and explanation (reuse existing function)
	code, explanation = ExtractCodeAndExplanation(response)

	// If no analysis found, try to extract from beginning
	if analysis == "" {
		lines := strings.Split(response, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "```") && !strings.HasPrefix(line, "ПРЕДЛОЖЕНИЯ") {
				analysis = line
				break
			}
		}
	}

	return analysis, suggestedActions, code, explanation
}
