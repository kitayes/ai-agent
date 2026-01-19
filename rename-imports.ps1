# Rename all arcgis-ai-assistant references to qgis-ai-assistant in .go files

Write-Host "Renaming import paths in Go files..." -ForegroundColor Green

$files = Get-ChildItem -Path . -Filter "*.go" -Recurse

$count = 0
foreach ($file in $files) {
    $content = Get-Content $file.FullName -Raw
    if ($content -match 'arcgis-ai-assistant/') {
        $newContent = $content -replace 'arcgis-ai-assistant/', 'qgis-ai-assistant/'
        Set-Content -Path $file.FullName -Value $newContent -NoNewline
        Write-Host "Updated: $($file.FullName)" -ForegroundColor Yellow
        $count++
    }
}

Write-Host "`nUpdated $count files" -ForegroundColor Green
Write-Host "`nNext steps:" -ForegroundColor Cyan
Write-Host "1. Run: go mod tidy"
Write-Host "2. Test: go run cmd/server/main.go"
