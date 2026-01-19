package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"qgis-ai-assistant/internal/datasources"
	"qgis-ai-assistant/internal/llm"
	"qgis-ai-assistant/internal/models"
)

// DataSearchHandler handles searching for available datasets
type DataSearchHandler struct {
	llmClient *llm.Client
	sources   map[string]datasources.DataSource
}

// NewDataSearchHandler creates a new data search handler
func NewDataSearchHandler(llmClient *llm.Client) *DataSearchHandler {
	return &DataSearchHandler{
		llmClient: llmClient,
		sources: map[string]datasources.DataSource{
			"osm": datasources.NewOSMDataSource(""),
			// Add more sources here as they're implemented
		},
	}
}

// DataSearchRequest represents a request to search for data
type DataSearchRequest struct {
	Prompt  string          `json:"prompt"`
	Context *models.Context `json:"context,omitempty"`
}

// DataSearchResponse represents the response
type DataSearchResponse struct {
	Source      string                `json:"source"`
	Datasets    []datasources.DataSet `json:"datasets"`
	Explanation string                `json:"explanation"`
	Error       string                `json:"error,omitempty"`
}

// Handle processes data search requests
func (h *DataSearchHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DataSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("=== DATA SEARCH REQUEST ===")
	log.Printf("Prompt: %s", req.Prompt)

	// Step 1: Ask AI which data source to use
	recommendation, err := h.selectDataSource(req.Prompt, req.Context)
	if err != nil {
		log.Printf("Error selecting data source: %v", err)
		resp := DataSearchResponse{
			Error: fmt.Sprintf("Failed to select data source: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	log.Printf("AI recommended source: %s", recommendation.Source)
	log.Printf("Reasoning: %s", recommendation.Reasoning)

	// Step 2: Get the appropriate data source
	source, ok := h.sources[recommendation.Source]
	if !ok {
		resp := DataSearchResponse{
			Error: fmt.Sprintf("Data source '%s' not available", recommendation.Source),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Step 3: Search for datasets
	datasets, err := source.Search(recommendation.SearchParams)
	if err != nil {
		log.Printf("Error searching datasets: %v", err)
		resp := DataSearchResponse{
			Source:      recommendation.Source,
			Error:       fmt.Sprintf("Failed to search: %v", err),
			Explanation: recommendation.Reasoning,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	log.Printf("Found %d datasets", len(datasets))

	// Return results
	resp := DataSearchResponse{
		Source:      recommendation.Source,
		Datasets:    datasets,
		Explanation: recommendation.Reasoning,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// selectDataSource asks AI which data source to use
func (h *DataSearchHandler) selectDataSource(prompt string, context *models.Context) (*DataSourceRecommendation, error) {
	// Build prompt for AI
	aiPrompt := fmt.Sprintf(`Пользователь запрашивает геоданные: "%s"

Доступные источники данных:
1. "osm" - OpenStreetMap (векторные данные: здания, дороги, POI, природные объекты)
2. "sentinel" - Спутниковые снимки Sentinel (пока не реализовано)
3. "geoportal_kz" - Геопорталы Казахстана (пока не реализовано)

Определи:
1. Какой источник использовать? (выбери из доступных: "osm")
2. Какие параметры поиска нужны?

Для OSM можешь задать:
- tags: {"building": "*"} для всех зданий, {"amenity": "school"} для школ, {"highway": "*"} для дорог
- keywords: ["название"] для поиска по имени

ВАЖНО: Определи bounding box для запроса. Используй известные координаты:
- Павлодар: 52.3°N, 76.95°E
- Павлодарская область: примерно 51.5-54.0°N, 75.0-80.0°E

ФОРМАТ ОТВЕТА (только JSON, без лишнего текста):
{
  "source": "osm",
  "bbox": {
    "minLat": 52.2,
    "minLon": 76.8,
    "maxLat": 52.4,
    "maxLon": 77.1
  },
  "tags": {"amenity": "school"},
  "reasoning": "Краткое объяснение почему выбран этот источник"
}`, prompt)

	// Call Gemini
	resp, err := h.llmClient.GenerateSimpleResponse(aiPrompt)
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	var recommendation DataSourceRecommendation
	if err := json.Unmarshal([]byte(resp), &recommendation); err != nil {
		// Try to extract JSON from response
		start := -1
		end := -1
		for i, ch := range resp {
			if ch == '{' && start == -1 {
				start = i
			}
			if ch == '}' {
				end = i + 1
			}
		}
		if start >= 0 && end > start {
			jsonPart := resp[start:end]
			if err := json.Unmarshal([]byte(jsonPart), &recommendation); err != nil {
				return nil, fmt.Errorf("failed to parse AI response: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to parse AI response: %w", err)
		}
	}

	// Convert to SearchParams
	recommendation.SearchParams = datasources.SearchParams{
		BoundingBox: recommendation.BBox,
		Tags:        recommendation.Tags,
		Keywords:    recommendation.Keywords,
		MaxResults:  100,
	}

	return &recommendation, nil
}

// DataSourceRecommendation represents AI's recommendation
type DataSourceRecommendation struct {
	Source       string                   `json:"source"`
	BBox         *datasources.BBox        `json:"bbox,omitempty"`
	Tags         map[string]string        `json:"tags,omitempty"`
	Keywords     []string                 `json:"keywords,omitempty"`
	Reasoning    string                   `json:"reasoning"`
	SearchParams datasources.SearchParams `json:"-"` // Filled by handler
}

// DataFetchHandler handles downloading datasets
type DataFetchHandler struct {
	sources   map[string]datasources.DataSource
	outputDir string
}

// NewDataFetchHandler creates a new data fetch handler
func NewDataFetchHandler(outputDir string) *DataFetchHandler {
	return &DataFetchHandler{
		sources: map[string]datasources.DataSource{
			"osm": datasources.NewOSMDataSource(""),
		},
		outputDir: outputDir,
	}
}

// DataFetchRequest represents a download request
type DataFetchRequest struct {
	Dataset datasources.DataSet `json:"dataset"`
}

// DataFetchResponse represents the download response
type DataFetchResponse struct {
	Success   bool   `json:"success"`
	FilePath  string `json:"filePath,omitempty"`
	LayerName string `json:"layerName,omitempty"`
	Error     string `json:"error,omitempty"`
}

// Handle processes data fetch requests
func (h *DataFetchHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DataFetchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("=== DATA FETCH REQUEST ===")
	log.Printf("Dataset: %s from %s", req.Dataset.Title, req.Dataset.Source)

	// Get data source
	source, ok := h.sources[req.Dataset.Source]
	if !ok {
		resp := DataFetchResponse{
			Success: false,
			Error:   fmt.Sprintf("Unknown data source: %s", req.Dataset.Source),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Generate output filename
	filename := fmt.Sprintf("%s_%s.geojson",
		req.Dataset.Source,
		req.Dataset.ID)
	outputPath := filepath.Join(h.outputDir, filename)

	// Download
	if err := source.Download(req.Dataset, outputPath); err != nil {
		log.Printf("Error downloading dataset: %v", err)
		resp := DataFetchResponse{
			Success: false,
			Error:   fmt.Sprintf("Download failed: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	log.Printf("Downloaded to: %s", outputPath)

	// Return success
	resp := DataFetchResponse{
		Success:   true,
		FilePath:  outputPath,
		LayerName: req.Dataset.Title,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
