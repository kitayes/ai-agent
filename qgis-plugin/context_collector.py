# -*- coding: utf-8 -*-
"""
Context Collector for QGIS
Collects project metadata, layer information, and spatial context
"""

import datetime
from qgis.core import (
    QgsProject,
    QgsVectorLayer,
    QgsRasterLayer,
    QgsWkbTypes,
    QgsMapLayer
)


class ContextCollector:
    """Collects full QGIS project context for AI"""

    def __init__(self):
        self.project = QgsProject.instance()

    def collect_full_context(self):
        """Collect complete project context including layers, CRS, extent"""
        
        context = {
            "project": self._collect_project_info(),
            "layers": self._collect_layers_info(),
            "timestamp": datetime.datetime.now().isoformat()
        }
        
        # Add active layer if exists
        from qgis.utils import iface
        if iface and iface.activeLayer():
            context["activeLayer"] = iface.activeLayer().name()
        
        # Add map extent if available
        if iface and iface.mapCanvas():
            extent = iface.mapCanvas().extent()
            context["mapExtent"] = {
                "xMin": extent.xMinimum(),
                "yMin": extent.yMinimum(),
                "xMax": extent.xMaximum(),
                "yMax": extent.yMaximum(),
                "scale": iface.mapCanvas().scale()
            }
        
        return context

    def _collect_project_info(self):
        """Collect basic project information"""
        
        project_info = {
            "name": self.project.baseName() or "Untitled Project",
            "path": self.project.fileName() or "",
            "spatialReference": "",
            "units": ""
        }
        
        # Get project CRS
        crs = self.project.crs()
        if crs.isValid():
            project_info["spatialReference"] = crs.authid()  # e.g. "EPSG:4326"
            project_info["units"] = self._get_units_name(crs.mapUnits())
        
        return project_info

    def _collect_layers_info(self):
        """Collect information about all layers in the project"""
        
        layers_info = []
        
        for layer_id, layer in self.project.mapLayers().items():
            layer_info = self._collect_single_layer_info(layer)
            if layer_info:
                layers_info.append(layer_info)
        
        return layers_info

    def _collect_single_layer_info(self, layer):
        """Collect detailed information about a single layer"""
        
        if not layer or not layer.isValid():
            return None
        
        layer_info = {
            "name": layer.name(),
            "type": self._get_layer_type_name(layer.type()),
            "isVisible": layer.isSpatial() if hasattr(layer, 'isSpatial') else True,
            "isEditable": layer.isEditable() if hasattr(layer, 'isEditable') else False
        }
        
        # CRS information
        if layer.isSpatial():
            crs = layer.crs()
            if crs.isValid():
                layer_info["spatialReference"] = crs.authid()
        
        # Vector layer specific info
        if isinstance(layer, QgsVectorLayer):
            layer_info.update(self._collect_vector_layer_info(layer))
        
        # Raster layer specific info
        elif isinstance(layer, QgsRasterLayer):
            layer_info.update(self._collect_raster_layer_info(layer))
        
        return layer_info

    def _collect_vector_layer_info(self, layer):
        """Collect vector layer specific information"""
        
        info = {
            "geometryType": self._get_geometry_type_name(layer.geometryType()),
            "featureCount": layer.featureCount(),
            "fields": []
        }
        
        # Collect field information
        for field in layer.fields():
            field_info = {
                "name": field.name(),
                "type": field.typeName(),
                "length": field.length(),
                "nullable": not field.constraints().constraints() & field.constraints().ConstraintNotNull
            }
            
            if field.alias():
                field_info["alias"] = field.alias()
            
            info["fields"].append(field_info)
        
        # Extent
        extent = layer.extent()
        if extent and not extent.isEmpty():
            info["extent"] = {
                "xMin": extent.xMinimum(),
                "yMin": extent.yMinimum(),
                "xMax": extent.xMaximum(),
                "yMax": extent.yMaximum()
            }
        
        # Data source (if not too long)
        source = layer.source()
        if source and len(source) < 500:
            info["dataSource"] = source
        
        return info

    def _collect_raster_layer_info(self, layer):
        """Collect raster layer specific information"""
        
        info = {
            "geometryType": "Raster",
            "width": layer.width(),
            "height": layer.height(),
            "bandCount": layer.bandCount()
        }
        
        # Extent
        extent = layer.extent()
        if extent and not extent.isEmpty():
            info["extent"] = {
                "xMin": extent.xMinimum(),
                "yMin": extent.yMinimum(),
                "xMax": extent.xMaximum(),
                "yMax": extent.yMaximum()
            }
        
        # Data source
        source = layer.source()
        if source and len(source) < 500:
            info["dataSource"] = source
        
        return info

    def _get_geometry_type_name(self, geom_type):
        """Convert QgsWkbTypes geometry type to readable name"""
        
        type_map = {
            QgsWkbTypes.PointGeometry: "Point",
            QgsWkbTypes.LineGeometry: "LineString",
            QgsWkbTypes.PolygonGeometry: "Polygon",
            QgsWkbTypes.UnknownGeometry: "Unknown",
            QgsWkbTypes.NullGeometry: "Null"
        }
        
        return type_map.get(geom_type, "Unknown")

    def _get_layer_type_name(self, layer_type):
        """Convert QgsMapLayer type to readable name"""
        
        type_map = {
            QgsMapLayer.VectorLayer: "Vector",
            QgsMapLayer.RasterLayer: "Raster",
            QgsMapLayer.PluginLayer: "Plugin",
            QgsMapLayer.MeshLayer: "Mesh",
            QgsMapLayer.VectorTileLayer: "VectorTile",
            QgsMapLayer.AnnotationLayer: "Annotation",
            QgsMapLayer.PointCloudLayer: "PointCloud"
        }
        
        return type_map.get(layer_type, "Unknown")

    def _get_units_name(self, units):
        """Convert map units enum to readable name"""
        
        from qgis.core import QgsUnitTypes
        
        unit_map = {
            QgsUnitTypes.DistanceMeters: "meters",
            QgsUnitTypes.DistanceKilometers: "kilometers",
            QgsUnitTypes.DistanceFeet: "feet",
            QgsUnitTypes.DistanceNauticalMiles: "nautical miles",
            QgsUnitTypes.DistanceYards: "yards",
            QgsUnitTypes.DistanceMiles: "miles",
            QgsUnitTypes.DistanceDegrees: "degrees",
            QgsUnitTypes.DistanceUnknownUnit: "unknown"
        }
        
        return unit_map.get(units, "unknown")
