"""
Context Collector for ArcGIS AI Assistant

This module collects comprehensive metadata about the ArcGIS project,
including layers, fields, spatial references, and current map state.
"""

import arcpy
import json
from typing import Dict, List, Optional


class ContextCollector:
    """Collects and structures ArcGIS project context"""
    
    def __init__(self):
        self.project = None
        try:
            self.project = arcpy.mp.ArcGISProject("CURRENT")
        except Exception as e:
            arcpy.AddWarning(f"Could not access current project: {e}")
    
    def collect_full_context(self) -> Dict:
        """
        Collect complete project context including:
        - Project info
        - All layers with metadata
        - Active layer
        - Current map extent
        """
        context = {
            "project": self.get_project_info(),
            "layers": self.get_all_layers(),
            "activeLayer": self.get_active_layer(),
            "mapExtent": self.get_map_extent(),
            "timestamp": arcpy.time.strftime("%Y-%m-%dT%H:%M:%S")
        }
        
        return context
    
    def get_project_info(self) -> Dict:
        """Get basic project information"""
        if not self.project:
            return {
                "name": "Unknown",
                "path": "",
                "spatialReference": "Unknown"
            }
        
        try:
            # Get first map's spatial reference
            sr = "Unknown"
            if self.project.listMaps():
                first_map = self.project.listMaps()[0]
                sr = first_map.spatialReference.name if first_map.spatialReference else "Unknown"
            
            return {
                "name": self.project.metadata.title or "Untitled Project",
                "path": self.project.filePath,
                "spatialReference": sr,
                "defaultDatabase": self.project.defaultGeodatabase
            }
        except Exception as e:
            arcpy.AddWarning(f"Error getting project info: {e}")
            return {"name": "Error", "path": "", "spatialReference": "Unknown"}
    
    def get_all_layers(self) -> List[Dict]:
        """Get metadata for all layers in all maps"""
        all_layers = []
        
        if not self.project:
            return all_layers
        
        try:
            for map_obj in self.project.listMaps():
                for layer in map_obj.listLayers():
                    layer_info = self.get_layer_details(layer)
                    if layer_info:
                        all_layers.append(layer_info)
        except Exception as e:
            arcpy.AddWarning(f"Error collecting layers: {e}")
        
        return all_layers
    
    def get_layer_details(self, layer) -> Optional[Dict]:
        """Get detailed metadata for a single layer"""
        try:
            # Skip group layers
            if layer.isGroupLayer:
                return None
            
            layer_info = {
                "name": layer.name,
                "type": self._get_layer_type(layer),
                "isVisible": layer.visible,
                "isEditable": False
            }
            
            # Get feature layer specific info
            if layer.supports("DATASOURCE"):
                layer_info["dataSource"] = layer.dataSource
            
            # Get feature count and geometry type for feature layers
            if hasattr(layer, 'dataSource') and layer.supports("DEFINITIONQUERY"):
                try:
                    desc = arcpy.Describe(layer)
                    
                    if hasattr(desc, 'shapeType'):
                        layer_info["geometryType"] = desc.shapeType
                    
                    if hasattr(desc, 'spatialReference'):
                        layer_info["spatialReference"] = desc.spatialReference.name
                    
                    # Get feature count
                    result = arcpy.management.GetCount(layer)
                    layer_info["featureCount"] = int(result[0])
                    
                    # Get extent
                    if hasattr(desc, 'extent'):
                        extent = desc.extent
                        layer_info["extent"] = {
                            "xMin": extent.XMin,
                            "yMin": extent.YMin,
                            "xMax": extent.XMax,
                            "yMax": extent.YMax
                        }
                    
                    # Get fields
                    layer_info["fields"] = self.get_layer_fields(layer)
                    
                    # Check if editable
                    layer_info["isEditable"] = hasattr(desc, 'canVersion')
                    
                except Exception as e:
                    arcpy.AddWarning(f"Could not get details for layer {layer.name}: {e}")
                    layer_info["featureCount"] = 0
            else:
                layer_info["featureCount"] = 0
                layer_info["spatialReference"] = "Unknown"
            
            return layer_info
            
        except Exception as e:
            arcpy.AddWarning(f"Error processing layer: {e}")
            return None
    
    def get_layer_fields(self, layer) -> List[Dict]:
        """Get field information for a layer"""
        fields = []
        
        try:
            for field in arcpy.ListFields(layer):
                # Skip system fields
                if field.name.upper() in ['OBJECTID', 'SHAPE', 'SHAPE_LENGTH', 'SHAPE_AREA', 'GLOBALID']:
                    continue
                
                field_info = {
                    "name": field.name,
                    "type": field.type,
                    "alias": field.aliasName or field.name,
                    "length": field.length if hasattr(field, 'length') else 0,
                    "nullable": field.isNullable
                }
                fields.append(field_info)
        except Exception as e:
            arcpy.AddWarning(f"Could not get fields: {e}")
        
        return fields[:10]  # Limit to first 10 fields to avoid huge payloads
    
    def get_active_layer(self) -> Optional[str]:
        """Get the currently active/selected layer"""
        if not self.project:
            return None
        
        try:
            active_view = self.project.activeView
            if hasattr(active_view, 'map'):
                # Get selected layers
                selected = active_view.map.listLayers()
                if selected:
                    return selected[0].name
        except Exception as e:
            arcpy.AddWarning(f"Could not get active layer: {e}")
        
        return None
    
    def get_map_extent(self) -> Optional[Dict]:
        """Get current map view extent"""
        if not self.project:
            return None
        
        try:
            active_view = self.project.activeView
            if hasattr(active_view, 'camera'):
                extent = active_view.camera.getExtent()
                return {
                    "xMin": extent.XMin,
                    "yMin": extent.YMin,
                    "xMax": extent.XMax,
                    "yMax": extent.YMax,
                    "scale": active_view.camera.scale
                }
        except Exception as e:
            arcpy.AddWarning(f"Could not get map extent: {e}")
        
        return None
    
    def _get_layer_type(self, layer) -> str:
        """Determine layer type"""
        if hasattr(layer, 'isFeatureLayer') and layer.isFeatureLayer:
            return "FeatureLayer"
        elif hasattr(layer, 'isRasterLayer') and layer.isRasterLayer:
            return "RasterLayer"
        elif hasattr(layer, 'isWebLayer') and layer.isWebLayer:
            return "WebLayer"
        else:
            return "Layer"
    
    def to_json(self) -> str:
        """Convert context to JSON string"""
        context = self.collect_full_context()
        return json.dumps(context, indent=2, ensure_ascii=False)


# Quick test function
def test_collector():
    """Test the context collector"""
    collector = ContextCollector()
    context = collector.collect_full_context()
    
    arcpy.AddMessage("=" * 50)
    arcpy.AddMessage("PROJECT CONTEXT")
    arcpy.AddMessage("=" * 50)
    arcpy.AddMessage(f"Project: {context['project']['name']}")
    arcpy.AddMessage(f"Layers found: {len(context['layers'])}")
    
    for layer in context['layers']:
        arcpy.AddMessage(f"  - {layer['name']} ({layer['geometryType']}, {layer['featureCount']} features)")
    
    return context


if __name__ == "__main__":
    test_collector()
