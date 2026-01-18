import arcpy
import pythonaddins
import json
import urllib.request
import urllib.error
import sys
import os

# Add the addon directory to path to import context_collector
addon_dir = os.path.dirname(__file__)
if addon_dir not in sys.path:
    sys.path.insert(0, addon_dir)

from context_collector import ContextCollector

SERVER_URL = "http://localhost:8080"

class AIAssistantTool(object):
    """Implementation for AI Assistant tool - Autonomous GIS Engineer"""
    def __init__(self):
        self.enabled = True
        self.checked = False
        self.context_collector = ContextCollector()
    
    def onClick(self):
        """Called when the tool button is clicked"""
        try:
            user_input = pythonaddins.MessageBox(
                "Введите команду для AI Assistant:\n\n(Например: 'Посчитай школы', 'Создай буфер 500м вокруг рек')",
                "ArcGIS AI Assistant - Autonomous GIS Engineer",
                0
            )
            
            if not user_input:
                return
            
            arcpy.AddMessage("=" * 60)
            arcpy.AddMessage(f"Запрос: {user_input}")
            arcpy.AddMessage("=" * 60)
            
            # Collect context
            arcpy.AddMessage("Сбор контекста проекта...")
            context = self.context_collector.collect_full_context()
            
            # Show context summary
            arcpy.AddMessage(f"Проект: {context['project']['name']}")
            arcpy.AddMessage(f"Доступно слоев: {len(context['layers'])}")
            for layer in context['layers'][:5]:  # Show first 5
                arcpy.AddMessage(f"  - {layer['name']} ({layer.get('geometryType', 'N/A')}, {layer['featureCount']} объектов)")
            if len(context['layers']) > 5:
                arcpy.AddMessage(f"  ... и еще {len(context['layers']) - 5} слоев")
            
            arcpy.AddMessage("\nОтправка запроса в AI...")
            
            code, explanation, warnings = self.send_to_ai(user_input, context)
            
            if code:
                arcpy.AddMessage("\n" + "=" * 60)
                arcpy.AddMessage("AI ОТВЕТ:")
                arcpy.AddMessage("=" * 60)
                arcpy.AddMessage(f"Объяснение: {explanation}")
                
                if warnings:
                    for warning in warnings:
                        arcpy.AddWarning(f"⚠️ {warning}")
                
                arcpy.AddMessage("\nГенерированный код:")
                arcpy.AddMessage("-" * 60)
                arcpy.AddMessage(code)
                arcpy.AddMessage("-" * 60)
                
                # Ask for confirmation
                confirm = pythonaddins.MessageBox(
                    f"Выполнить сгенерированный код?\n\n{explanation}\n\nКод:\n{code[:200]}...",
                    "Подтверждение выполнения",
                    1  # Yes/No
                )
                
                if confirm == "Yes":
                    arcpy.AddMessage("\nВыполнение кода...")
                    try:
                        # Execute in controlled environment
                        exec_globals = {
                            'arcpy': arcpy,
                            '__builtins__': __builtins__
                        }
                        exec(code, exec_globals)
                        arcpy.AddMessage("\n✅ Код успешно выполнен!")
                    except Exception as e:
                        error_msg = str(e)
                        arcpy.AddError(f"❌ Ошибка выполнения: {error_msg}")
                        
                        # Try to regenerate
                        arcpy.AddMessage("\nПопытка исправления ошибки...")
                        fixed_code = self.regenerate_code(user_input, code, error_msg, context)
                        
                        if fixed_code:
                            arcpy.AddMessage("AI исправил код. Повторная попытка...")
                            try:
                                exec(fixed_code, exec_globals)
                                arcpy.AddMessage("✅ Исправленный код выполнен успешно!")
                            except Exception as e2:
                                arcpy.AddError(f"❌ Ошибка повторного выполнения: {str(e2)}")
                else:
                    arcpy.AddMessage("Выполнение отменено пользователем")
            else:
                arcpy.AddError("Не удалось получить код от AI")
                
        except Exception as e:
            arcpy.AddError(f"❌ Критическая ошибка: {str(e)}")
            import traceback
            arcpy.AddError(traceback.format_exc())
    
    def send_to_ai(self, prompt, context):
        """Send request to AI backend with full context"""
        try:
            url = f"{SERVER_URL}/api/generate"
            
            payload = {
                "prompt": prompt,
                "context": context
            }
            
            data = json.dumps(payload, ensure_ascii=False).encode('utf-8')
            
            req = urllib.request.Request(
                url,
                data=data,
                headers={'Content-Type': 'application/json; charset=utf-8'}
            )
            
            with urllib.request.urlopen(req, timeout=60) as response:
                result = json.loads(response.read().decode('utf-8'))
                
                if 'error' in result and result['error']:
                    arcpy.AddError(f"AI Error: {result['error']}")
                    return None, None, None
                
                return (
                    result.get('code'),
                    result.get('explanation'),
                    result.get('warnings', [])
                )
                
        except urllib.error.URLError as e:
            arcpy.AddError(f"❌ Ошибка подключения к серверу: {str(e)}")
            arcpy.AddError(f"Убедитесь, что сервер запущен на {SERVER_URL}")
            return None, None, None
        except Exception as e:
            arcpy.AddError(f"❌ Ошибка отправки запроса: {str(e)}")
            import traceback
            arcpy.AddError(traceback.format_exc())
            return None, None, None
    
    def regenerate_code(self, original_prompt, failed_code, error_message, context, attempt=1):
        """Try to regenerate fixed code after error"""
        if attempt > 3:
            arcpy.AddError("Превышено максимальное количество попыток исправления")
            return None
        
        try:
            url = f"{SERVER_URL}/api/regenerate"
            
            payload = {
                "originalPrompt": original_prompt,
                "failedCode": failed_code,
                "errorMessage": error_message,
                "context": context,
                "attempt": attempt
            }
            
            data = json.dumps(payload, ensure_ascii=False).encode('utf-8')
            
            req = urllib.request.Request(
                url,
                data=data,
                headers={'Content-Type': 'application/json; charset=utf-8'}
            )
            
            with urllib.request.urlopen(req, timeout=60) as response:
                result = json.loads(response.read().decode('utf-8'))
                
                if 'error' in result and result['error']:
                    arcpy.AddError(f"Ошибка регенерации: {result['error']}")
                    return None
                
                arcpy.AddMessage(f"Объяснение исправления: {result.get('explanation', 'N/A')}")
                return result.get('code')
                
        except Exception as e:
            arcpy.AddError(f"Ошибка регенерации кода: {str(e)}")
            return None

