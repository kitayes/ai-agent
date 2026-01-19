# Phase 2: Autonomous Data Fetching  - Testing Guide

## What's Implemented

✅ **OpenStreetMap Integration**
- Overpass API client
- GeoJSON conversion
- AI-driven source selection
- HTTP endpoints

## Testing Steps

### 1. Start Server

```powershell
# Make sure you're in project directory
cd c:\Users\user\.gemini\antigravity\scratch\arcgis-ai-assistant

# Configure UTF-8
.\configure-utf8.ps1

# Start server
go run cmd/server/main.go
```

### 2. Test Data Search

**Example 1: Schools in Pavlodar**

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/data/search" `
  -Method POST `
  -ContentType "application/json" `
  -Body '{"prompt":"Найди все школы в Павлодаре"}'
```

Expected response:
```json
{
  "source": "osm",
  "datasets": [{
    "id": "osm-...",
    "title": "OSM: amenity=school",
    "source": "osm",
    "format": "GeoJSON"
  }],
  "explanation": "OSM selected because user requested POI data (schools)"
}
```

**Example 2: Roads**

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/data/search" `
  -Method POST `
  -ContentType "application/json" `
  -Body '{"prompt":"Скачай дороги OpenStreetMap для Павлодара"}'
```

**Example 3: Buildings**

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/data/search" `
  -Method POST `
  -ContentType "application/json" `
  -Body '{"prompt":"Загрузи все здания в радиусе 2км от центра Павлодара"}'
```

### 3. Test Data Fetch

After getting dataset from search, fetch it:

```powershell
# Save search result
$searchResult = Invoke-RestMethod -Uri "http://localhost:8080/api/data/search" `
  -Method POST `
  -ContentType "application/json" `
  -Body '{"prompt":"Найти школы в Павлодаре"}'

# Extract first dataset
$dataset = $searchResult.datasets[0]

# Fetch the data
Invoke-RestMethod -Uri "http://localhost:8080/api/data/fetch" `
  -Method POST `
  -ContentType "application/json" `
  -Body ($dataset | ConvertTo-Json -Depth 10)
```

Expected response:
```json
{
  "success": true,
  "filePath": "./downloads/osm_osm-1234567.geojson",
  "layerName": "OSM: amenity=school"
}
```

### 4. Verify Downloaded File

```powershell
# Check downloads directory
Get-ChildItem ./downloads

# View GeoJSON content
Get-Content ./downloads/<filename>.geojson | Select-Object -First 50
```

## Known Limitations (Current Implementation)

1. ⚠️ **Only OSM implemented** - Sentinel and Kazakhstan portals coming later
2. ⚠️ **No QGIS integration yet** - files downloaded but not auto-imported
3  ⚠️ **Bbox estimation** - AI estimates bounding boxes, may need refinement
4. ⚠️ **Rate limiting** - Overpass API has rate limits, avoid spamming

## Next Steps

After successful testing:

1. ✅ Test OSM integration
2. [ ] Implement Sentinel satellite data source
3. [ ] Add Kazakhstan geo-portals
4. [ ] Update QGIS plugin for automatic import
5. [ ] Add /api/data/sources endpoint

## Troubleshooting

### Error: "bounding box is required"

AI failed to determine bbox. Try more specific prompt:
```
"Найди школы в Павлодаре (52.3°N, 77.0°E)"
```

### Error: Overpass API timeout

Query too large or API overloaded. Try:
- Smaller area
- More specific tags
- Wait and retry

### Empty results

Check:
- Bbox coordinates correct?
- Tags valid? (e.g., `amenity=school`)
- Area has OSM data?

## Example Queries for Kazakhstan

```bash
# Pavlodar schools
"Найди школы в Павлодаре"

# Roads in region
"Скачай главные дороги Павлодарской области"

# Rivers
"Загрузи Иртыш из OpenStreetMap"

# Buildings
"Получи все здания в центре Павлодара"

# Parks
"Найди парки и скверы в Павлодаре"
```

## File Structure

```
qgis-ai-assistant/
├── internal/
│   ├── datasources/
│   │   ├── interface.go      # DataSource interface
│   │   └── osm.go            # OSM implementation
│   ├── handlers/
│   │   └── data_handlers.go  # Search & fetch handlers
│   └── llm/
│       └── client.go         # +GenerateSimpleResponse
├── downloads/                 # Downloaded files (created automatically)
└── .env                      # Config
```
