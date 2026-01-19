package datasources

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// OSMDataSource implements DataSource for OpenStreetMap via Overpass API
type OSMDataSource struct {
	overpassURL string
	client      *http.Client
}

// NewOSMDataSource creates a new OpenStreetMap data source
func NewOSMDataSource(overpassURL string) *OSMDataSource {
	if overpassURL == "" {
		overpassURL = "https://overpass-api.de/api/interpreter"
	}

	return &OSMDataSource{
		overpassURL: overpassURL,
		client: &http.Client{
			Timeout: 120 * time.Second, // OSM queries can take time
		},
	}
}

// Name returns the data source name
func (o *OSMDataSource) Name() string {
	return "OpenStreetMap"
}

// Search finds OSM features matching the parameters
func (o *OSMDataSource) Search(params SearchParams) ([]DataSet, error) {
	if params.BoundingBox == nil {
		return nil, fmt.Errorf("bounding box is required for OSM search")
	}

	// Build Overpass QL query
	query := o.buildQuery(params.BoundingBox, params.Tags, params.Keywords)

	// Execute query
	data, err := o.executeQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute OSM query: %w", err)
	}

	// Parse response
	var response OverpassResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse OSM response: %w", err)
	}

	// Convert to datasets
	datasets := []DataSet{
		{
			ID:    fmt.Sprintf("osm-%d", time.Now().Unix()),
			Title: o.buildTitle(params),
			Description: fmt.Sprintf("OpenStreetMap data for bbox (%.2f,%.2f,%.2f,%.2f)",
				params.BoundingBox.MinLat, params.BoundingBox.MinLon,
				params.BoundingBox.MaxLat, params.BoundingBox.MaxLon),
			Source:      "osm",
			BoundingBox: params.BoundingBox,
			Date:        time.Now(),
			Format:      "GeoJSON",
			Size:        int64(len(data)),
			Metadata: map[string]interface{}{
				"elements_count": len(response.Elements),
				"query":          query,
			},
		},
	}

	return datasets, nil
}

// Download downloads OSM data and saves as GeoJSON
func (o *OSMDataSource) Download(dataset DataSet, outputPath string) error {
	// Extract query from metadata
	query, ok := dataset.Metadata["query"].(string)
	if !ok {
		return fmt.Errorf("no query found in dataset metadata")
	}

	// Execute query
	data, err := o.executeQuery(query)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	// Convert Overpass JSON to GeoJSON
	geoJSON, err := o.convertToGeoJSON(data)
	if err != nil {
		return fmt.Errorf("failed to convert to GeoJSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, geoJSON, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// GetMetadata retrieves metadata (OSM doesn't have per-dataset metadata)
func (o *OSMDataSource) GetMetadata(datasetID string) (*Metadata, error) {
	return &Metadata{
		ID:          datasetID,
		Title:       "OpenStreetMap Data",
		Description: "OpenStreetMap crowdsourced geographic data",
		Source:      "osm",
		Format:      "GeoJSON",
		License:     "ODbL (Open Database License)",
		Attribution: "© OpenStreetMap contributors",
	}, nil
}

// buildQuery constructs an Overpass QL query
func (o *OSMDataSource) buildQuery(bbox *BBox, tags map[string]string, keywords []string) string {
	// Overpass bbox format: south,west,north,east
	bboxStr := fmt.Sprintf("%.6f,%.6f,%.6f,%.6f",
		bbox.MinLat, bbox.MinLon, bbox.MaxLat, bbox.MaxLon)

	var parts []string

	// Add tag-based queries
	if len(tags) > 0 {
		for key, value := range tags {
			if value == "*" {
				parts = append(parts, fmt.Sprintf("node[\"%s\"](%s);", key, bboxStr))
				parts = append(parts, fmt.Sprintf("way[\"%s\"](%s);", key, bboxStr))
				parts = append(parts, fmt.Sprintf("relation[\"%s\"](%s);", key, bboxStr))
			} else {
				parts = append(parts, fmt.Sprintf("node[\"%s\"=\"%s\"](%s);", key, value, bboxStr))
				parts = append(parts, fmt.Sprintf("way[\"%s\"=\"%s\"](%s);", key, value, bboxStr))
				parts = append(parts, fmt.Sprintf("relation[\"%s\"=\"%s\"](%s);", key, value, bboxStr))
			}
		}
	}

	// Add keyword-based queries (search in name)
	if len(keywords) > 0 {
		for _, keyword := range keywords {
			parts = append(parts, fmt.Sprintf("node[\"name\"~\"%s\",i](%s);", keyword, bboxStr))
			parts = append(parts, fmt.Sprintf("way[\"name\"~\"%s\",i](%s);", keyword, bboxStr))
		}
	}

	// Default: get all features if no specific queries
	if len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("node(%s);", bboxStr))
		parts = append(parts, fmt.Sprintf("way(%s);", bboxStr))
	}

	// Construct full query
	query := "[out:json][timeout:90];\n(\n  "
	query += strings.Join(parts, "\n  ")
	query += "\n);\nout geom;"

	return query
}

// executeQuery sends query to Overpass API
func (o *OSMDataSource) executeQuery(query string) ([]byte, error) {
	data := url.Values{}
	data.Set("data", query)

	req, err := http.NewRequest("POST", o.overpassURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("overpass API returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// convertToGeoJSON converts Overpass JSON to GeoJSON
func (o *OSMDataSource) convertToGeoJSON(data []byte) ([]byte, error) {
	var response OverpassResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	// Build GeoJSON
	geoJSON := GeoJSON{
		Type:     "FeatureCollection",
		Features: make([]Feature, 0, len(response.Elements)),
	}

	for _, elem := range response.Elements {
		feature := o.elementToFeature(elem)
		if feature != nil {
			geoJSON.Features = append(geoJSON.Features, *feature)
		}
	}

	return json.MarshalIndent(geoJSON, "", "  ")
}

// elementToFeature converts OSM element to GeoJSON feature
func (o *OSMDataSource) elementToFeature(elem Element) *Feature {
	if elem.Lat == 0 && elem.Lon == 0 && len(elem.Geometry) == 0 {
		return nil // Skip elements without geometry
	}

	feature := &Feature{
		Type:       "Feature",
		Properties: elem.Tags,
	}

	// Add OSM metadata
	if feature.Properties == nil {
		feature.Properties = make(map[string]interface{})
	}
	feature.Properties["osm_id"] = elem.ID
	feature.Properties["osm_type"] = elem.Type

	// Set geometry based on type
	switch elem.Type {
	case "node":
		feature.Geometry = Geometry{
			Type:        "Point",
			Coordinates: []interface{}{elem.Lon, elem.Lat},
		}

	case "way":
		if len(elem.Geometry) > 0 {
			coords := make([]interface{}, len(elem.Geometry))
			for i, pt := range elem.Geometry {
				coords[i] = []float64{pt.Lon, pt.Lat}
			}
			feature.Geometry = Geometry{
				Type:        "LineString",
				Coordinates: coords,
			}
		}

	case "relation":
		// Simplified - just use first member's geometry
		if len(elem.Members) > 0 {
			// Would need more complex handling for multipolygons
			return nil
		}
	}

	return feature
}

// buildTitle creates a descriptive title from search params
func (o *OSMDataSource) buildTitle(params SearchParams) string {
	if len(params.Tags) > 0 {
		// Use first relevant tag
		for key, value := range params.Tags {
			if value == "*" {
				return fmt.Sprintf("OSM: All %s", key)
			}
			return fmt.Sprintf("OSM: %s=%s", key, value)
		}
	}

	if len(params.Keywords) > 0 {
		return fmt.Sprintf("OSM: %s", strings.Join(params.Keywords, ", "))
	}

	return "OSM: Geographic data"
}

// OverpassResponse represents Overpass API response
type OverpassResponse struct {
	Version   float64   `json:"version"`
	Generator string    `json:"generator"`
	Elements  []Element `json:"elements"`
}

// Element represents an OSM element
type Element struct {
	Type     string                 `json:"type"`
	ID       int64                  `json:"id"`
	Lat      float64                `json:"lat,omitempty"`
	Lon      float64                `json:"lon,omitempty"`
	Tags     map[string]interface{} `json:"tags,omitempty"`
	Geometry []GeometryPoint        `json:"geometry,omitempty"`
	Members  []Member               `json:"members,omitempty"`
}

// GeometryPoint represents a point in way geometry
type GeometryPoint struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// Member represents a relation member
type Member struct {
	Type string `json:"type"`
	Ref  int64  `json:"ref"`
	Role string `json:"role"`
}

// GeoJSON structures
type GeoJSON struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

type Feature struct {
	Type       string                 `json:"type"`
	Geometry   Geometry               `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

type Geometry struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates"`
}
