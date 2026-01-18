package llm

import (
	"context"
	"fmt"
	"strings"

	"arcgis-ai-assistant/internal/models"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Client struct {
	model *genai.GenerativeModel
	ctx   context.Context
}

func NewClient(ctx context.Context, apiKey string) (*Client, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	model := client.GenerativeModel("gemini-1.5-pro")
	model.SetTemperature(0.2)
	model.SetTopP(0.8)
	model.SetTopK(40)
	model.SetMaxOutputTokens(2048)

	return &Client{
		model: model,
		ctx:   ctx,
	}, nil
}

// GenerateCodeWithContext generates ArcPy code with project context
func (c *Client) GenerateCodeWithContext(userPrompt string, projectContext *models.Context) (code, explanation string, usedLayers, warnings []string, err error) {
	fullPrompt := BuildPromptWithContext(userPrompt, projectContext)

	resp, err := c.model.GenerateContent(c.ctx, genai.Text(fullPrompt))
	if err != nil {
		return "", "", nil, nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", "", nil, nil, fmt.Errorf("empty response from Gemini")
	}

	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	code, explanation = ExtractCodeAndExplanation(responseText)
	usedLayers = ExtractUsedLayers(code, projectContext)
	warnings = GenerateWarnings(code, projectContext)

	return code, explanation, usedLayers, warnings, nil
}

// RegenerateCode attempts to fix failed code
func (c *Client) RegenerateCode(originalPrompt, failedCode, errorMessage string, projectContext *models.Context, attempt int) (code, explanation string, usedLayers, warnings []string, err error) {
	fullPrompt := BuildRegenerationPrompt(originalPrompt, failedCode, errorMessage, projectContext, attempt)

	resp, err := c.model.GenerateContent(c.ctx, genai.Text(fullPrompt))
	if err != nil {
		return "", "", nil, nil, fmt.Errorf("failed to regenerate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", "", nil, nil, fmt.Errorf("empty response from Gemini")
	}

	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	code, explanation = ExtractCodeAndExplanation(responseText)
	usedLayers = ExtractUsedLayers(code, projectContext)
	warnings = GenerateWarnings(code, projectContext)

	return code, explanation, usedLayers, warnings, nil
}

// ExtractUsedLayers identifies which layers are referenced in the code
func ExtractUsedLayers(code string, context *models.Context) []string {
	if context == nil {
		return []string{}
	}

	var used []string
	codeLower := strings.ToLower(code)

	for _, layer := range context.Layers {
		layerNameLower := strings.ToLower(layer.Name)
		if strings.Contains(codeLower, layerNameLower) || strings.Contains(codeLower, fmt.Sprintf("\"%s\"", layer.Name)) {
			used = append(used, layer.Name)
		}
	}

	return used
}

// GenerateWarnings checks for potential issues in generated code
func GenerateWarnings(code string, context *models.Context) []string {
	var warnings []string

	// Check for layer references without context
	if context == nil || len(context.Layers) == 0 {
		if strings.Contains(code, "SelectLayer") || strings.Contains(code, "Buffer") {
			warnings = append(warnings, "Код использует слои, но контекст проекта недоступен")
		}
	}

	// Check for potentially long operations
	if strings.Contains(code, "Clip") || strings.Contains(code, "Union") || strings.Contains(code, "Intersect") {
		warnings = append(warnings, "Операция может занять несколько минут на больших датасетах")
	}

	// Check for in-memory workspace usage
	if strings.Contains(code, "in_memory") {
		warnings = append(warnings, "Используется временное хранилище in_memory")
	}

	return warnings
}

// AnalyzeMapScreenshot analyzes a map screenshot using Gemini Vision
func (c *Client) AnalyzeMapScreenshot(imageBytes []byte, userPrompt string, projectContext *models.Context) (analysis string, suggestedActions []string, code, explanation string, warnings []string, err error) {
	prompt := BuildVisionPrompt(userPrompt, projectContext)

	// Create image part for Gemini
	imagePart := genai.ImageData("png", imageBytes)

	resp, err := c.model.GenerateContent(c.ctx, genai.Text(prompt), imagePart)
	if err != nil {
		return "", nil, "", "", nil, fmt.Errorf("failed to analyze screenshot: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", nil, "", "", nil, fmt.Errorf("empty response from Gemini Vision")
	}

	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	// Parse vision response
	analysis, suggestedActions, code, explanation = ParseVisionResponse(responseText)
	warnings = []string{}

	// If code was generated, validate it
	if code != "" {
		usedLayers := ExtractUsedLayers(code, projectContext)
		warnings = GenerateWarnings(code, projectContext)

		// Add used layers info
		if len(usedLayers) > 0 {
			analysis += fmt.Sprintf("\n\nИспользуемые слои: %v", usedLayers)
		}
	}

	return analysis, suggestedActions, code, explanation, warnings, nil
}
