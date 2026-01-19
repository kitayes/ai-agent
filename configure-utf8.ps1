# PowerShell UTF-8 Configuration for QGIS AI Assistant
# Run this before starting the server to fix Russian text encoding

Write-Host "Configuring PowerShell for UTF-8..." -ForegroundColor Green

# Set console output encoding to UTF-8
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

# Set console input encoding to UTF-8
[Console]::InputEncoding = [System.Text.Encoding]::UTF8

# Set default encoding for PowerShell session
$PSDefaultParameterValues['*:Encoding'] = 'utf8'

# Verify settings
Write-Host "`nCurrent Encoding Settings:" -ForegroundColor Cyan
Write-Host "Output Encoding: $([Console]::OutputEncoding.EncodingName)"
Write-Host "Input Encoding: $([Console]::InputEncoding.EncodingName)"
Write-Host "PowerShell Default: UTF-8"

Write-Host "`nUTF-8 configured successfully!" -ForegroundColor Green
Write-Host "Now you can run: go run cmd/server/main.go" -ForegroundColor Yellow
Write-Host ""
Write-Host "Test with:" -ForegroundColor Yellow
Write-Host 'Invoke-RestMethod -Uri "http://localhost:8080/api/generate" -Method POST -ContentType "application/json" -Body ''{"prompt":"Покажи сообщение Привет"}''' -ForegroundColor Gray
