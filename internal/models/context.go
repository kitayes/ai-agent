package models

import "time"

// Context represents the full ArcGIS project context
type Context struct {
	Project     ProjectInfo `json:"project"`
	Layers      []LayerInfo `json:"layers"`
	ActiveLayer string      `json:"activeLayer,omitempty"`
	MapExtent   *MapExtent  `json:"mapExtent,omitempty"`
	Timestamp   time.Time   `json:"timestamp"`
}

// ProjectInfo contains project metadata
type ProjectInfo struct {
	Name             string `json:"name"`
	Path             string `json:"path"`
	SpatialReference string `json:"spatialReference"`
	DefaultDatabase  string `json:"defaultDatabase,omitempty"`
}

// LayerInfo contains detailed layer metadata
type LayerInfo struct {
	Name             string       `json:"name"`
	Type             string       `json:"type"`
	GeometryType     string       `json:"geometryType,omitempty"`
	FeatureCount     int          `json:"featureCount"`
	DataSource       string       `json:"dataSource,omitempty"`
	Fields           []FieldInfo  `json:"fields,omitempty"`
	SpatialReference string       `json:"spatialReference"`
	Extent           *LayerExtent `json:"extent,omitempty"`
	IsVisible        bool         `json:"isVisible"`
	IsEditable       bool         `json:"isEditable"`
}

// FieldInfo describes a field in a layer
type FieldInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Alias    string `json:"alias,omitempty"`
	Length   int    `json:"length,omitempty"`
	Nullable bool   `json:"nullable"`
}

// LayerExtent defines spatial bounds of a layer
type LayerExtent struct {
	XMin float64 `json:"xMin"`
	YMin float64 `json:"yMin"`
	XMax float64 `json:"xMax"`
	YMax float64 `json:"yMax"`
}

// MapExtent defines the current map view extent
type MapExtent struct {
	XMin  float64 `json:"xMin"`
	YMin  float64 `json:"yMin"`
	XMax  float64 `json:"xMax"`
	YMax  float64 `json:"yMax"`
	Scale float64 `json:"scale,omitempty"`
}

// GenerateRequest is the request structure for code generation
type GenerateRequest struct {
	Prompt  string   `json:"prompt"`
	Context *Context `json:"context,omitempty"`
}

// GenerateResponse is the response structure
type GenerateResponse struct {
	Code        string   `json:"code"`
	Explanation string   `json:"explanation"`
	UsedLayers  []string `json:"usedLayers,omitempty"`
	Warnings    []string `json:"warnings,omitempty"`
	Error       string   `json:"error,omitempty"`
}

// RegenerateRequest for error correction
type RegenerateRequest struct {
	OriginalPrompt string   `json:"originalPrompt"`
	FailedCode     string   `json:"failedCode"`
	ErrorMessage   string   `json:"errorMessage"`
	Context        *Context `json:"context,omitempty"`
	Attempt        int      `json:"attempt"`
}
