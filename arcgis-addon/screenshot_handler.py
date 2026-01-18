"""
Screenshot Handler for ArcGIS AI Assistant

Captures map screenshots and prepares them for vision analysis
"""

import arcpy
import os
import tempfile
from datetime import datetime


class ScreenshotHandler:
    """Handles screenshot capture from ArcGIS Pro"""
    
    def __init__(self):
        self.project = None
        try:
            self.project = arcpy.mp.ArcGISProject("CURRENT")
        except Exception as e:
            arcpy.AddWarning(f"Could not access current project: {e}")
    
    def capture_current_map(self, width=1920, height=1080):
        """
        Capture screenshot of the current map view
        
        Args:
            width: Screenshot width in pixels
            height: Screenshot height in pixels
            
        Returns:
            str: Path to the captured screenshot
        """
        if not self.project:
            arcpy.AddError("No active project found")
            return None
        
        try:
            # Get active map view
            active_map = self.project.activeMap
            if not active_map:
                arcpy.AddError("No active map found")
                return None
            
            # Create temp filename with timestamp
            timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
            temp_dir = tempfile.gettempdir()
            screenshot_path = os.path.join(temp_dir, f"arcgis_screenshot_{timestamp}.png")
            
            arcpy.AddMessage(f"Capturing screenshot ({width}x{height})...")
            
            # Export map to PNG
            # Note: This requires ArcGIS Pro 2.8+
            layout = None
            for lyt in self.project.listLayouts():
                if lyt.name == active_map.name or len(self.project.listLayouts()) == 1:
                    layout = lyt
                    break
            
            if layout:
                layout.exportToPNG(screenshot_path, resolution=96, width=width, height=height)
            else:
                # Fallback: try to export map directly
                arcpy.AddMessage("Creating temporary layout for export...")
                # This is a simplified approach - in production you'd create a proper layout
                arcpy.AddWarning("No layout found - using map export")
                
                # Create a simple export using arcpy
                # Note: This may require additional setup
                screenshot_path = self._export_map_view(active_map, screenshot_path, width, height)
            
            if os.path.exists(screenshot_path):
                file_size = os.path.getsize(screenshot_path) / 1024  # KB
                arcpy.AddMessage(f"Screenshot saved: {screenshot_path} ({file_size:.1f} KB)")
                return screenshot_path
            else:
                arcpy.AddError("Screenshot file was not created")
                return None
                
        except Exception as e:
            arcpy.AddError(f"Error capturing screenshot: {e}")
            import traceback
            arcpy.AddError(traceback.format_exc())
            return None
    
    def _export_map_view(self, map_obj, output_path, width, height):
        """
        Fallback method to export map view
        
        This is a simplified version. In production, you would:
        1. Create a temporary layout
        2. Set the layout size
        3. Export the layout
        4. Clean up
        """
        try:
            # Get the project's default layout or create one
            project = self.project
            
            # Try to find or create a layout
            layouts = project.listLayouts()
            if layouts:
                layout = layouts[0]
            else:
                arcpy.AddWarning("Cannot export without a layout - please create one in your project")
                return None
            
            # Export
            layout.exportToPNG(output_path, resolution=96)
            return output_path
            
        except Exception as e:
            arcpy.AddError(f"Export fallback failed: {e}")
            return None
    
    def get_screenshot_metadata(self, screenshot_path):
        """Get metadata about the screenshot"""
        if not os.path.exists(screenshot_path):
            return None
        
        return {
            "path": screenshot_path,
            "size_kb": os.path.getsize(screenshot_path) / 1024,
            "timestamp": datetime.fromtimestamp(os.path.getmtime(screenshot_path)).isoformat(),
            "exists": True
        }


def test_screenshot():
    """Test screenshot capture"""
    handler = ScreenshotHandler()
    
    arcpy.AddMessage("Testing screenshot capture...")
    screenshot_path = handler.capture_current_map(width=1280, height=720)
    
    if screenshot_path:
        metadata = handler.get_screenshot_metadata(screenshot_path)
        arcpy.AddMessage(f"Success! Screenshot metadata: {metadata}")
        return screenshot_path
    else:
        arcpy.AddError("Screenshot capture failed")
        return None


if __name__ == "__main__":
    test_screenshot()
