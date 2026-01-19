package datasources

import "time"

// DataSource is the interface for all external data sources
type DataSource interface {
	// Search for available datasets
	Search(params SearchParams) ([]DataSet, error)

	// Download dataset to local file
	Download(dataset DataSet, outputPath string) error

	// GetMetadata retrieves detailed information about a dataset
	GetMetadata(datasetID string) (*Metadata, error)

	// Name returns the data source name
	Name() string
}

// SearchParams defines parameters for searching datasets
type SearchParams struct {
	BoundingBox *BBox             `json:"boundingBox,omitempty"`
	StartDate   time.Time         `json:"startDate,omitempty"`
	EndDate     time.Time         `json:"endDate,omitempty"`
	MaxResults  int               `json:"maxResults,omitempty"`
	Keywords    []string          `json:"keywords,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"` // For OSM queries
}

// BBox defines a geographic bounding box
type BBox struct {
	MinLat float64 `json:"minLat"` // South
	MinLon float64 `json:"minLon"` // West
	MaxLat float64 `json:"maxLat"` // North
	MaxLon float64 `json:"maxLon"` // East
}

// DataSet represents a single dataset from any source
type DataSet struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Source      string                 `json:"source"` // "sentinel", "osm", "geoportal_kz"
	BoundingBox *BBox                  `json:"boundingBox,omitempty"`
	Date        time.Time              `json:"date,omitempty"`
	Format      string                 `json:"format"` // "GeoTIFF", "Shapefile", "GeoJSON"
	DownloadURL string                 `json:"downloadUrl,omitempty"`
	Size        int64                  `json:"size,omitempty"`       // in bytes
	CloudCover  float64                `json:"cloudCover,omitempty"` // for satellite imagery
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Metadata contains detailed information about a dataset
type Metadata struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Source      string                 `json:"source"`
	BoundingBox *BBox                  `json:"boundingBox,omitempty"`
	Date        time.Time              `json:"date,omitempty"`
	Format      string                 `json:"format"`
	Size        int64                  `json:"size,omitempty"`
	License     string                 `json:"license,omitempty"`
	Attribution string                 `json:"attribution,omitempty"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// Common bounding boxes for Kazakhstan regions
var (
	// Pavlodar Region (approximately)
	PavlodarBBox = &BBox{
		MinLat: 51.5,
		MinLon: 75.0,
		MaxLat: 54.0,
		MaxLon: 80.0,
	}

	// Pavlodar City (approximately)
	PavlodarCityBBox = &BBox{
		MinLat: 52.2,
		MinLon: 76.8,
		MaxLat: 52.4,
		MaxLon: 77.1,
	}

	// Kazakhstan (full country)
	KazakhstanBBox = &BBox{
		MinLat: 40.5,
		MinLon: 46.5,
		MaxLat: 55.5,
		MaxLon: 87.5,
	}
)

// NewBBoxFromCenter creates a bounding box from center point and radius
func NewBBoxFromCenter(lat, lon, radiusKm float64) *BBox {
	// Approximate conversion: 1 degree ≈ 111 km at equator
	latDelta := radiusKm / 111.0
	lonDelta := radiusKm / (111.0 * cosApprox(lat))

	return &BBox{
		MinLat: lat - latDelta,
		MinLon: lon - lonDelta,
		MaxLat: lat + latDelta,
		MaxLon: lon + lonDelta,
	}
}

// cosApprox approximates cosine for latitude (in degrees)
func cosApprox(lat float64) float64 {
	// Simple approximation for cos(lat) in degrees
	latRad := lat * 0.017453292519943295     // deg to rad
	return 0.5 + 0.5*(1.0-latRad*latRad/2.0) // Taylor approximation
}

// Contains checks if a point is within the bounding box
func (b *BBox) Contains(lat, lon float64) bool {
	return lat >= b.MinLat && lat <= b.MaxLat &&
		lon >= b.MinLon && lon <= b.MaxLon
}

// Area returns approximate area in square kilometers
func (b *BBox) Area() float64 {
	widthKm := (b.MaxLon - b.MinLon) * 111.0 * cosApprox((b.MinLat+b.MaxLat)/2.0)
	heightKm := (b.MaxLat - b.MinLat) * 111.0
	return widthKm * heightKm
}
