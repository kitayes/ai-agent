# -*- coding: utf-8 -*-
"""
AI Assistant for GIS - Main Plugin Class
Autonomous GIS Engineer for QGIS powered by Gemini AI
"""

import os
import json
import urllib.request
import urllib.error
from qgis.PyQt.QtCore import QSettings, QTranslator, QCoreApplication, Qt
from qgis.PyQt.QtGui import QIcon
from qgis.PyQt.QtWidgets import QAction, QInputDialog, QMessageBox
from qgis.core import QgsMessageLog, Qgis, QgsProject
from .context_collector import ContextCollector

SERVER_URL = "http://localhost:8080"


class AIAssistant:
    """QGIS Plugin Implementation - Autonomous GIS Engineer"""

    def __init__(self, iface):
        """Constructor.
        
        :param iface: An interface instance that will be passed to this class
            which provides the hook by which you can manipulate the QGIS
            application at run time.
        :type iface: QgsInterface
        """
        self.iface = iface
        self.plugin_dir = os.path.dirname(__file__)
        
        # Initialize locale
        locale = QSettings().value('locale/userLocale')[0:2]
        locale_path = os.path.join(
            self.plugin_dir,
            'i18n',
            'AIAssistant_{}.qm'.format(locale))

        if os.path.exists(locale_path):
            self.translator = QTranslator()
            self.translator.load(locale_path)
            QCoreApplication.installTranslator(self.translator)

        # Declare instance attributes
        self.actions = []
        self.menu = self.tr(u'&AI Assistant')
        
        # Context collector
        self.context_collector = ContextCollector()

    def tr(self, message):
        """Get the translation for a string using Qt translation API."""
        return QCoreApplication.translate('AIAssistant', message)

    def add_action(
        self,
        icon_path,
        text,
        callback,
        enabled_flag=True,
        add_to_menu=True,
        add_to_toolbar=True,
        status_tip=None,
        whats_this=None,
        parent=None):
        """Add a toolbar icon to the toolbar.
        
        :param icon_path: Path to the icon for this action
        :param text: Text that should be shown in menu items for this action
        :param callback: Function to be called when the action is triggered
        :param enabled_flag: A flag indicating if the action should be enabled by default
        :param add_to_menu: Flag indicating whether the action should also be added to the menu
        :param add_to_toolbar: Flag indicating whether the action should also be added to the toolbar
        :param status_tip: Optional text to show in a popup when mouse pointer hovers over the action
        :param parent: Parent widget for the new action
        :param whats_this: Optional text to show in the status bar when the mouse pointer hovers over the action
        
        :returns: The action that was created
        :rtype: QAction
        """

        icon = QIcon(icon_path)
        action = QAction(icon, text, parent)
        action.triggered.connect(callback)
        action.setEnabled(enabled_flag)

        if status_tip is not None:
            action.setStatusTip(status_tip)

        if whats_this is not None:
            action.setWhatsThis(whats_this)

        if add_to_toolbar:
            # Adds plugin icon to Plugins toolbar
            self.iface.addToolBarIcon(action)

        if add_to_menu:
            self.iface.addPluginToMenu(
                self.menu,
                action)

        self.actions.append(action)

        return action

    def initGui(self):
        """Create the menu entries and toolbar icons inside the QGIS GUI."""

        icon_path = os.path.join(self.plugin_dir, 'icon.png')
        self.add_action(
            icon_path,
            text=self.tr(u'AI Assistant - Autonomous GIS Engineer'),
            callback=self.run,
            parent=self.iface.mainWindow(),
            status_tip=self.tr(u'Launch AI Assistant for natural language GIS commands'))

    def unload(self):
        """Removes the plugin menu item and icon from QGIS GUI."""
        for action in self.actions:
            self.iface.removePluginMenu(
                self.tr(u'&AI Assistant'),
                action)
            self.iface.removeToolBarIcon(action)

    def run(self):
        """Run method that performs all the real work"""
        
        # Get user input
        text, ok = QInputDialog.getText(
            self.iface.mainWindow(),
            "AI Assistant - Autonomous GIS Engineer",
            "Введите команду для AI:\n\n"
            "(Например: 'Посчитай школы', 'Создай буфер 500м вокруг рек', "
            "'Загрузи данные Sentinel для Павлодара')"
        )
        
        if not ok or not text:
            return
            
        QgsMessageLog.logMessage("=" * 60, "AI Assistant", Qgis.Info)
        QgsMessageLog.logMessage(f"Запрос: {text}", "AI Assistant", Qgis.Info)
        QgsMessageLog.logMessage("=" * 60, "AI Assistant", Qgis.Info)
        
        # Collect context
        QgsMessageLog.logMessage("Сбор контекста проекта...", "AI Assistant", Qgis.Info)
        context = self.context_collector.collect_full_context()
        
        # Show context summary
        QgsMessageLog.logMessage(f"Проект: {context['project']['name']}", "AI Assistant", Qgis.Info)
        QgsMessageLog.logMessage(f"Доступно слоев: {len(context['layers'])}", "AI Assistant", Qgis.Info)
        
        for i, layer in enumerate(context['layers'][:5]):  # Show first 5
            geom_type = layer.get('geometryType', 'N/A')
            count = layer.get('featureCount', 0)
            QgsMessageLog.logMessage(
                f"  - {layer['name']} ({geom_type}, {count} объектов)",
                "AI Assistant",
                Qgis.Info
            )
        
        if len(context['layers']) > 5:
            QgsMessageLog.logMessage(
                f"  ... и еще {len(context['layers']) - 5} слоев",
                "AI Assistant",
                Qgis.Info
            )
        
        QgsMessageLog.logMessage("\nОтправка запроса в AI...", "AI Assistant", Qgis.Info)
        
        # Send to AI
        try:
            code, explanation, warnings = self.send_to_ai(text, context)
            
            if code:
                QgsMessageLog.logMessage("\n" + "=" * 60, "AI Assistant", Qgis.Info)
                QgsMessageLog.logMessage("AI ОТВЕТ:", "AI Assistant", Qgis.Info)
                QgsMessageLog.logMessage("=" * 60, "AI Assistant", Qgis.Info)
                QgsMessageLog.logMessage(f"Объяснение: {explanation}", "AI Assistant", Qgis.Info)
                
                # Show warnings
                if warnings:
                    for warning in warnings:
                        QgsMessageLog.logMessage(f"⚠️ {warning}", "AI Assistant", Qgis.Warning)
                
                QgsMessageLog.logMessage("\nГенерированный код:", "AI Assistant", Qgis.Info)
                QgsMessageLog.logMessage("-" * 60, "AI Assistant", Qgis.Info)
                QgsMessageLog.logMessage(code, "AI Assistant", Qgis.Info)
                QgsMessageLog.logMessage("-" * 60, "AI Assistant", Qgis.Info)
                
                # Ask for confirmation
                reply = QMessageBox.question(
                    self.iface.mainWindow(),
                    "Подтверждение выполнения",
                    f"Выполнить сгенерированный код?\n\n{explanation}\n\nКод:\n{code[:300]}...",
                    QMessageBox.Yes | QMessageBox.No,
                    QMessageBox.No
                )
                
                if reply == QMessageBox.Yes:
                    QgsMessageLog.logMessage("\nВыполнение кода...", "AI Assistant", Qgis.Info)
                    self.execute_code(code, text, context)
                else:
                    QgsMessageLog.logMessage("Выполнение отменено пользователем", "AI Assistant", Qgis.Info)
            else:
                QgsMessageLog.logMessage("❌ Не удалось получить код от AI", "AI Assistant", Qgis.Critical)
                
        except Exception as e:
            QgsMessageLog.logMessage(f"❌ Критическая ошибка: {str(e)}", "AI Assistant", Qgis.Critical)
            import traceback
            QgsMessageLog.logMessage(traceback.format_exc(), "AI Assistant", Qgis.Critical)

    def execute_code(self, code, original_prompt, context):
        """Execute generated PyQGIS code safely"""
        try:
            # Prepare execution environment
            from qgis.core import *
            from qgis import processing
            
            exec_globals = {
                'qgis': __import__('qgis'),
                'processing': processing,
                'QgsProject': QgsProject,
                'QgsMessageLog': QgsMessageLog,
                'Qgis': Qgis,
                'iface': self.iface,
                '__builtins__': __builtins__
            }
            
            # Import all from qgis.core
            from qgis.core import *
            for name in dir():
                if name.startswith('Qgs'):
                    exec_globals[name] = eval(name)
            
            # Execute code
            exec(code, exec_globals)
            
            QgsMessageLog.logMessage("\n✅ Код успешно выполнен!", "AI Assistant", Qgis.Success)
            self.iface.messageBar().pushMessage(
                "AI Assistant",
                "Код успешно выполнен!",
                level=Qgis.Success,
                duration=5
            )
            
        except Exception as e:
            error_msg = str(e)
            QgsMessageLog.logMessage(f"❌ Ошибка выполнения: {error_msg}", "AI Assistant", Qgis.Critical)
            
            # Try to regenerate
            QgsMessageLog.logMessage("\nПопытка исправления ошибки...", "AI Assistant", Qgis.Info)
            fixed_code = self.regenerate_code(original_prompt, code, error_msg, context)
            
            if fixed_code:
                QgsMessageLog.logMessage("AI исправил код. Повторная попытка...", "AI Assistant", Qgis.Info)
                try:
                    exec(fixed_code, exec_globals)
                    QgsMessageLog.logMessage("✅ Исправленный код выполнен успешно!", "AI Assistant", Qgis.Success)
                    self.iface.messageBar().pushMessage(
                        "AI Assistant",
                        "Исправленный код выполнен успешно!",
                        level=Qgis.Success,
                        duration=5
                    )
                except Exception as e2:
                    QgsMessageLog.logMessage(f"❌ Ошибка повторного выполнения: {str(e2)}", "AI Assistant", Qgis.Critical)
                    self.iface.messageBar().pushMessage(
                        "AI Assistant",
                        f"Ошибка выполнения: {str(e2)}",
                        level=Qgis.Critical,
                        duration=10
                    )

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
                    QgsMessageLog.logMessage(f"AI Error: {result['error']}", "AI Assistant", Qgis.Critical)
                    return None, None, None
                
                return (
                    result.get('code'),
                    result.get('explanation'),
                    result.get('warnings', [])
                )
                
        except urllib.error.URLError as e:
            QgsMessageLog.logMessage(f"❌ Ошибка подключения к серверу: {str(e)}", "AI Assistant", Qgis.Critical)
            QgsMessageLog.logMessage(f"Убедитесь, что сервер запущен на {SERVER_URL}", "AI Assistant", Qgis.Critical)
            self.iface.messageBar().pushMessage(
                "AI Assistant",
                f"Ошибка подключения к серверу на {SERVER_URL}",
                level=Qgis.Critical,
                duration=10
            )
            return None, None, None
        except Exception as e:
            QgsMessageLog.logMessage(f"❌ Ошибка отправки запроса: {str(e)}", "AI Assistant", Qgis.Critical)
            import traceback
            QgsMessageLog.logMessage(traceback.format_exc(), "AI Assistant", Qgis.Critical)
            return None, None, None

    def regenerate_code(self, original_prompt, failed_code, error_message, context, attempt=1):
        """Try to regenerate fixed code after error"""
        if attempt > 3:
            QgsMessageLog.logMessage("Превышено максимальное количество попыток исправления", "AI Assistant", Qgis.Critical)
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
                    QgsMessageLog.logMessage(f"Ошибка регенерации: {result['error']}", "AI Assistant", Qgis.Critical)
                    return None
                
                QgsMessageLog.logMessage(f"Объяснение исправления: {result.get('explanation', 'N/A')}", "AI Assistant", Qgis.Info)
                return result.get('code')
                
        except Exception as e:
            QgsMessageLog.logMessage(f"Ошибка регенерации кода: {str(e)}", "AI Assistant", Qgis.Critical)
            return None
