# ArcGIS AI Assistant - Инструкция по установке

## Шаг 1: Установка зависимостей Go

```powershell
cd C:\Users\user\.gemini\antigravity\scratch\arcgis-ai-assistant
go mod download
```

## Шаг 2: Настройка API ключа

1. Создайте файл `.env` из примера:
```powershell
copy .env.example .env
```

2. Откройте файл `.env` и добавьте ваш Gemini API ключ:
```
GEMINI_API_KEY=ваш_ключ_здесь
SERVER_PORT=8080
LOG_LEVEL=info
```

## Шаг 3: Запуск сервера

```powershell
go run cmd/server/main.go
```

Сервер должен запуститься на `http://localhost:8080`

## Шаг 4: Тестирование API

### Тест Echo:
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/echo" -Method POST -ContentType "application/json" -Body '{"message":"Hello"}'
```

### Тест Generate:
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/generate" -Method POST -ContentType "application/json" -Body '{"prompt":"Покажи сообщение Привет"}'
```

## Шаг 5: Создание Add-in для ArcGIS

```powershell
cd arcgis-addon
Compress-Archive -Path * -DestinationPath ..\ArcGISAIAssistant.zip -Force
Rename-Item ..\ArcGISAIAssistant.zip ..\ArcGISAIAssistant.esriaddin
```

## Шаг 6: Установка в ArcGIS Pro

1. Дважды кликните на файл `ArcGISAIAssistant.esriaddin`
2. Подтвердите установку
3. Перезапустите ArcGIS Pro
4. Найдите панель "AI Assistant" в интерфейсе

## Готово!

Теперь можно использовать AI Assistant в ArcGIS Pro.
