# QGIS Plugin Installation Guide

## Requirement

1. **QGIS Desktop 3.0+** (Рекомендуется QGIS 3.34 LTR)
   - Скачать: https://qgis.org/download/
   - Бесплатно и open-source

2. **Go Backend Server** (уже настроен в проекте)
   - Запускается из корня проекта: `go run cmd/server/main.go`

3. **Gemini API Key**
   - Получить на: https://makersuite.google.com/app/apikey
   - Добавить в `.env` файл

## Plugin Installation

### Метод 1: Через QGIS Plugin Manager (Рекомендуется)

1. Создайте ZIP архив из папки `qgis-plugin/`:
   ```powershell
   cd qgis-plugin
   Compress-Archive -Path * -DestinationPath ../ai_assistant.zip
   ```

2. Откройте QGIS

3. Перейдите: **Plugins → Manage and Install Plugins**

4. Нажмите **Install from ZIP**

5. Выберите `ai_assistant.zip`

6. Нажмите **Install Plugin**

7. Plugin появится в меню **Plugins → AI Assistant**

### Метод 2: Ручная установка

1. Найдите папку QGIS plugins:
   - Windows: `C:\Users\<USER>\AppData\Roaming\QGIS\QGIS3\profiles\default\python\plugins\`
   - Linux: `~/.local/share/QGIS/QGIS3/profiles/default/python/plugins/`
   - Mac: `~/Library/Application Support/QGIS/QGIS3/profiles/default/python/plugins/`

2. Создайте папку `ai_assistant` в этой директории

3. Скопируйте все файлы из `qgis-plugin/` в `ai_assistant/`

4. Перезапустите QGIS

5. Включите plugin: **Plugins → Manage and Install Plugins → Installed → AI Assistant** (поставьте галочку)

## Usage

1. **Запустите Go Backend**:
   ```bash
   cd qgis-ai-assistant
   go run cmd/server/main.go
   ```
   Сервер должен быть запущен на `http://localhost:8080`

2. **Откройте QGIS** с вашим проектом (например, с данными Павлодарской области)

3. **Нажмите на иконку AI Assistant** в панели инструментов
   - Или: Меню **Plugins → AI Assistant → AI Assistant - Autonomous GIS Engineer**

4. **Введите команду** на естественном языке:
   - "Посчитай количество школ"
   - "Создай буфер 500 метров вокруг рек"
   - "Выбери все здания в радиусе 1км от центра города"
   - "Загрузи спутниковый снимок Sentinel для Павлодара"

5. **AI сгенерирует PyQGIS код**, покажет объяснение и предупреждения

6. **Подтвердите выполнение** - код будет выполнен в QGIS

## Logging

Все логи отображаются в **Log Messages Panel** в QGIS:
- View → Panels → Log Messages
- Фильтр: "AI Assistant"

## Troubleshooting

### Plugin не появляется в меню

1. Проверьте, что plugin включен в Plugin Manager
2. Проверьте Log Messages на наличие ошибок загрузки
3. Убедитесь, что все файлы скопированы корректно

### Ошибка подключения к серверу

1. Проверьте, что Go backend запущен: `http://localhost:8080/health`
2. Проверьте, что порт 8080 не занят
3. Проверьте firewall settings

### Ошибка выполнения кода

1. Проверьте Log Messages для деталей ошибки
2. AI автоматически попытается исправить код
3. Если не помогло - переформулируйте запрос

### API Key ошибки

1. Проверьте `.env` файл в корне проекта
2. Убедитесь, что `GEMINI_API_KEY` установлен корректно
3. Проверьте квоты API на https://makersuite.google.com/

## Examples for Kazakhstan

### Павлодарская область

```
"Создай буфер 500м вокруг Иртыша"
"Посчитай школы в городе Павлодар"
"Выбери все здания ближе 1км от реки"
"Загрузи данные OSM для Павлодара"
```

### Системы координат

```
"Перепроецируй слой в EPSG:32643" (UTM Zone 43N для Павлодара)
"Преобразуй в WGS84"
```

### Анализ

```
"Найди пересечения между слоями дорог и зданий"
"Вычисли площадь всех полигонов"
"Создай тепловую карту плотности точек"
```

## Next Steps

После установки и тестирования базовой функциональности, можно:

1. Добавить автономный data fetching (Sentinel, OSM)
2. Настроить мониторинг территорий
3. Интегрировать дополнительные источники данных для Казахстана

## Support

- GitHub Issues: https://github.com/yourusername/qgis-ai-assistant/issues
- Documentation: См. `docs/` папку в проекте
