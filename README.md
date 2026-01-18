# QGIS AI Assistant - Автономный ГИС-Инженер для Казахстана

AI-powered autonomous GIS assistant for **QGIS** using Gemini Pro, optimized for Kazakhstan territory.

**Целевая аудитория:** Специалисты ГИС в Казахстане (Павлодарская область и другие регионы)

**Платформа:** QGIS 3.0+ (бесплатно и open-source)

## Архитектура

```
QGIS (PyQGIS) ←→ Go Backend (HTTP) ←→ Gemini AI
```

### Компоненты

1. **Go Backend** - HTTP сервер с Gemini интеграцией
2. **QGIS Plugin** - плагин для QGIS (PyQGIS)
3. **Gemini AI** - генерация PyQGIS скриптов

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

### 2. Установка QGIS Plugin

#### Создание .zip файла плагина:

1. Перейдите в каталог `qgis-plugin`
2. Заархивируйте все файлы в ZIP
3. Установите через QGIS Plugin Manager

Подробные инструкции: `qgis-plugin/INSTALL_PLUGIN.md`

#### Или командная строка:

```powershell
cd qgis-plugin
Compress-Archive -Path * -DestinationPath ../ai_assistant.zip
```

Затем в QGIS: **Plugins → Manage and Install Plugins → Install from ZIP**

### 3. Использование

1. Запустите Go сервер (должен работать на фоне)
2. Откройте ArcGIS Pro с проектом территории Казахстана
3. Найдите панель инструментов "AI Assistant"
4. Нажмите кнопку AI Assistant
5. Введите команду, например:
   - "Посчитай школы в Павлодаре"
   - "Создай буфер 500м вокруг Иртыша"
   - "Загрузи спутниковый снимок Павлодарской области"
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
