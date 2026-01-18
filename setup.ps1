# Quick Start Script
# Run this after setting up .env file

Write-Host "=== ArcGIS AI Assistant - Quick Start ===" -ForegroundColor Cyan

# Check if .env exists
if (-Not (Test-Path ".env")) {
    Write-Host "❌ .env file not found!" -ForegroundColor Red
    Write-Host "Please copy .env.example to .env and add your Gemini API key" -ForegroundColor Yellow
    Write-Host "`nRun: copy .env.example .env" -ForegroundColor White
    exit 1
}

# Check Go installation
Write-Host "`n[1/4] Checking Go installation..." -ForegroundColor Yellow
$goVersion = go version
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ $goVersion" -ForegroundColor Green
} else {
    Write-Host "❌ Go not installed" -ForegroundColor Red
    exit 1
}

# Download dependencies
Write-Host "`n[2/4] Downloading Go dependencies..." -ForegroundColor Yellow
go mod download
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Dependencies downloaded" -ForegroundColor Green
} else {
    Write-Host "❌ Failed to download dependencies" -ForegroundColor Red
    exit 1
}

# Build server
Write-Host "`n[3/4] Building server..." -ForegroundColor Yellow
go build -o bin/server.exe cmd/server/main.go
if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Server built successfully" -ForegroundColor Green
} else {
    Write-Host "❌ Build failed" -ForegroundColor Red
    exit 1
}

# Create .esriaddin
Write-Host "`n[4/4] Creating ArcGIS Add-in..." -ForegroundColor Yellow
Push-Location arcgis-addon
Compress-Archive -Path * -DestinationPath ..\ArcGISAIAssistant.zip -Force
Pop-Location
Rename-Item .\ArcGISAIAssistant.zip .\ArcGISAIAssistant.esriaddin -Force
if (Test-Path "ArcGISAIAssistant.esriaddin") {
    Write-Host "✅ Add-in created: ArcGISAIAssistant.esriaddin" -ForegroundColor Green
} else {
    Write-Host "❌ Failed to create add-in" -ForegroundColor Red
}

Write-Host "`n=== Setup Complete! ===" -ForegroundColor Cyan
Write-Host "`nNext steps:" -ForegroundColor Yellow
Write-Host "1. Run server: .\bin\server.exe" -ForegroundColor White
Write-Host "2. Install add-in: Double-click ArcGISAIAssistant.esriaddin" -ForegroundColor White
Write-Host "3. Open ArcGIS Pro and use AI Assistant tool" -ForegroundColor White
