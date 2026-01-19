# UTF-8 Encoding Fix Guide

## Problem

Windows PowerShell shows Russian text as mojibake (кракозябры):
- Expected: "Код сгенерирован успешно"
- Actual: "ÐÐ¾Ð´ ÑÐ³ÐµÐ½ÐµÑÐ¸ÑÐ¾Ð²Ð°Ð½ ÑÑÐ¿ÐµÑÐ½Ð¾"

## Solution

### Quick Fix (Recommended)

Run this **once per PowerShell session**:

```powershell
.\configure-utf8.ps1
```

Then start the server:

```powershell
go run cmd/server/main.go
```

### Manual Fix

If you prefer manual configuration:

```powershell
# Set console encoding
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8
[Console]::InputEncoding = [System.Text.Encoding]::UTF8

# Then start server
go run cmd/server/main.go
```

### Permanent Fix (Optional)

Add to your PowerShell profile for automatic UTF-8:

```powershell
# Edit profile
notepad $PROFILE

# Add these lines:
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8
[Console]::InputEncoding = [System.Text.Encoding]::UTF8
```

## Testing

After configuration, test with:

```powershell
# Start server
go run cmd/server/main.go

# In another PowerShell window (also run configure-utf8.ps1 first):
Invoke-RestMethod -Uri "http://localhost:8080/api/generate" `
  -Method POST `
  -ContentType "application/json; charset=utf-8" `
  -Body '{"prompt":"Покажи сообщение Привет"}'
```

You should see:
```
code        explanation
----        -----------
            Код сгенерирован успешно
```

## Notes

- This is a Windows PowerShell limitation, not a project issue
- Linux/Mac terminals handle UTF-8 natively
- Server logs will still be in English to avoid console issues
- API responses are properly UTF-8 encoded
