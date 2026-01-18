# -*- coding: utf-8 -*-
"""
AI Assistant for GIS - QGIS Plugin Initialization
"""

def classFactory(iface):
    """Load AIAssistant class from file ai_assistant.
    
    :param iface: A QGIS interface instance.
    :type iface: QgsInterface
    """
    from .ai_assistant import AIAssistant
    return AIAssistant(iface)
