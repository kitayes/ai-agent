# Start Server Script
# Runs the ArcGIS AI Assistant server

Write-Host "Starting ArcGIS AI Assistant Server..." -ForegroundColor Cyan

if (-Not (Test-Path ".env")) {
    Write-Host "❌ .env file not found!" -ForegroundColor Red
    Write-Host "Run setup.ps1 first" -ForegroundColor Yellow
    exit 1
}

if (Test-Path "bin/server.exe") {
    .\bin\server.exe
} else {
    Write-Host "Server binary not found, running from source..." -ForegroundColor Yellow
    go run cmd/server/main.go
}
