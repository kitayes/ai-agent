# ArcGIS AI Assistant

AI-powered assistant for ArcGIS Pro using Gemini and Go backend.

## Архитектура

```
ArcGIS Pro (Python) ←→ Go Backend (HTTP) ←→ Gemini AI
```

### Компоненты

1. **Go Backend** - HTTP сервер с Gemini интеграцией
2. **Python Add-in** - инструмент внутри ArcGIS Pro
3. **Gemini AI** - генерация ArcPy скриптов

## Быстрый старт

### 1. Настройка Go Backend

```powershell
cd arcgis-ai-assistant

# Создайте .env файл
copy .env.example .env

# Отредактируйте .env и добавьте ваш GEMINI_API_KEY
notepad .env

# Установите зависимости
go mod download

# Запустите сервер
go run cmd/server/main.go
```

Сервер запустится на `http://localhost:8080`

### 2. Установка Python Add-in

#### Создание .esriaddin файла:

1. Откройте ArcGIS Pro
2. Перейдите в каталог `arcgis-addon`
3. Заархивируйте все файлы в ZIP
4. Переименуйте расширение `.zip` в `.esriaddin`
5. Дважды кликните на `.esriaddin` для установки

#### Или используйте командную строку:

```powershell
cd arcgis-addon
Compress-Archive -Path * -DestinationPath ../ArcGISAIAssistant.zip
Rename-Item ../ArcGISAIAssistant.zip ../ArcGISAIAssistant.esriaddin
```

### 3. Использование

1. Запустите Go сервер (должен работать на фоне)
2. Откройте ArcGIS Pro
3. Найдите панель инструментов "AI Assistant"
4. Нажмите кнопку AI Assistant
5. Введите команду, например: "Покажи сообщение Привет"
6. AI сгенерирует и выполнит код

## API Endpoints

### POST `/api/echo`
Тестовый endpoint для проверки связи.

**Request:**
```json
{
  "message": "Hello"
}
```

**Response:**
```json
{
  "response": "Hello"
}
```

### POST `/api/generate`
Генерация ArcPy кода через Gemini.

**Request:**
```json
{
  "prompt": "Покажи сообщение Привет"
}
```

**Response:**
```json
{
  "code": "arcpy.AddMessage(\"Привет\")",
  "explanation": "Этот код выводит сообщение 'Привет' в окно геопроцессинга"
}
```

### GET `/health`
Health check endpoint.

## Тестирование

### Тест Echo endpoint:

```powershell
curl -X POST http://localhost:8080/api/echo `
  -H "Content-Type: application/json" `
  -d '{\"message\": \"Hello\"}'
```

### Тест Generate endpoint:

```powershell
curl -X POST http://localhost:8080/api/generate `
  -H "Content-Type: application/json" `
  -d '{\"prompt\": \"Покажи сообщение Привет\"}'
```

## Структура проекта

```
arcgis-ai-assistant/
├── cmd/
│   └── server/
│       └── main.go              # Точка входа
├── internal/
│   ├── config/
│   │   └── config.go            # Конфигурация
│   ├── server/
│   │   └── server.go            # HTTP сервер
│   ├── handlers/
│   │   ├── echo.go              # Echo endpoint
│   │   └── generate.go          # Generate endpoint
│   └── llm/
│       ├── client.go            # Gemini клиент
│       └── prompt.go            # Промпты
├── arcgis-addon/
│   ├── ai_tool.py               # Python add-in
│   └── Config.daml              # Конфигурация add-in
├── .env.example                 # Пример конфигурации
├── go.mod                       # Go зависимости
└── README.md                    # Документация
```

## Требования

- Go 1.21+
- ArcGIS Pro 3.0+
- Gemini API Key

## Следующие фазы

- **Phase 2**: Сбор метаданных слоев
- **Phase 3**: Сложный геопроцессинг (буферы, выборки)
- **Phase 4**: Многошаговые workflow

## Troubleshooting

### Ошибка подключения к серверу

Убедитесь, что:
1. Go сервер запущен (`go run cmd/server/main.go`)
2. Сервер доступен на `http://localhost:8080`
3. Нет блокировки firewall

### Ошибка API Key

Проверьте:
1. Файл `.env` существует
2. `GEMINI_API_KEY` установлен корректно
3. API key активен в Google AI Studio

## Лицензия

MIT
